"""
We use the objectstore to get the latest and
greatest of the stadsarchief xml!
"""

import os
import logging

from swiftclient.client import Connection

import datetime

from dateutil import parser

log = logging.getLogger('objectstore')

assert os.getenv('BEELDBANK_OBJECTSTORE_PASSWORD', 'need magixword')
assert os.getenv('XMLPARSER_DATA_PATH', 'where to store xml?')

OBJECTSTORE = dict(
    VERSION='2.0',
    AUTHURL='https://identity.stack.cloudvps.com/v2.0',
    TENANT_NAME='BGE000081 Beeldbank',
    TENANT_ID='2b4fd2c3fcc544d2a2f1c4256098e84d',
    USER=os.getenv('OBJECTSTORE_USER', 'beeldbank'),
    PASSWORD=os.getenv('BEELDBANK_OBJECTSTORE_PASSWORD'),
    REGION_NAME='NL',
)

DATA_DIR = os.getenv('XML_PARSER_DATA_PATH', '/app/data/')

logging.getLogger("requests").setLevel(logging.WARNING)
logging.getLogger("urllib3").setLevel(logging.WARNING)
logging.getLogger("swiftclient").setLevel(logging.WARNING)


store = OBJECTSTORE


EXPECTED_FILES = [
    '.xml',
]
SOURCE_DIR = 'Beeldbank XML'
IMAGE_DIR = 'Beeldbank Files'

# global..
os_conn = {}


def get_connection():

    options = {
        'tenant_id': store['TENANT_ID'],
        'region_name': store['REGION_NAME'],
        'endpoint_type': 'internalURL', }

    log.debug("Do we run local? set ENV LOCAL to something")

    if os.getenv('LOCAL', False):
        log.debug("We run LOCAL!")
        options.pop('endpoint_type')

    new_handelsregister_conn = Connection(
        authurl=store['AUTHURL'],
        user=store['USER'],
        key=store['PASSWORD'],
        tenant_name=store['TENANT_NAME'],
        auth_version=store['VERSION'],
        os_options=options)

    return new_handelsregister_conn


def get_store_object(object_meta_data):
    return os_conn.get_object(
        SOURCE_DIR, object_meta_data['name'])[1]


def get_full_container_list(conn, container, **kwargs):
    limit = 10000
    kwargs['limit'] = limit
    page = []

    seed = []

    _, page = conn.get_container(container, **kwargs)
    seed.extend(page)

    while len(page) == limit:
        # keep getting pages..
        kwargs['marker'] = seed[-1]['name']
        _, page = conn.get_container(container, **kwargs)
        seed.extend(page)

    return seed


def download_files(file_list):
    # Download the latest data
    for _, source_data_file in file_list:

        xml_file = source_data_file['name'].split('/')[-1]
        msg = 'Downloading: %s' % (xml_file)
        log.debug(msg)

        output_file = f'{DATA_DIR}/{xml_file}'

        if os.path.isfile(output_file):
            log.debug('SKIPPED: File already downloaded %s', output_file)
            continue

        new_data = get_store_object(source_data_file)

        # save output to file!
        with open(output_file, 'wb') as outputxml:
            outputxml.write(new_data)


def _get_latest_xml_files():
    """
    Download the expected files provided by mks / kpn
    """
    file_list = []

    meta_data = get_full_container_list(
        os_conn, SOURCE_DIR)

    for o_info in meta_data:
        for expected_file in EXPECTED_FILES:
            # if not expected continue.
            if not o_info['name'].endswith(expected_file):
                continue

            dt = parser.parse(o_info['last_modified'])
            now = datetime.datetime.now()

            delta = now - dt

            log.debug('AGE: %d %s', delta.days, expected_file)

            if delta.days > 50:
                log.error('DELEVERY IMPORTED FILES ARE TOO OLD!')
                raise ValueError

            log.debug('%s %s', expected_file, dt)
            file_list.append((dt, o_info))

    download_files(file_list)


def _get_full_imageslist():
    """
    Download an overview of all image files
    """

    full_list = get_full_container_list(os_conn, IMAGE_DIR)

    output_file = f'{DATA_DIR}/image_list.txt'

    if os.path.isfile(output_file):
        os.remove(output_file)

    # save output to file
    with open(output_file, 'wb') as outputfile:
        for line in full_list:
            outputfile.write(str.encode(line['name']+'\n'))


if __name__ == '__main__':
    logging.basicConfig(level=logging.DEBUG)
    os_conn = get_connection()
    _get_latest_xml_files()
    _get_full_imageslist()
