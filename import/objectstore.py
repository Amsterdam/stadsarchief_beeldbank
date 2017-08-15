import logging
import os
from functools import lru_cache

from swiftclient.client import Connection

log = logging.getLogger(__name__)

logging.getLogger("requests").setLevel(logging.WARNING)
logging.getLogger("urllib3").setLevel(logging.WARNING)
logging.getLogger("swiftclient").setLevel(logging.WARNING)

os_connect = {
    'auth_version': '2.0',
    'authurl': 'https://identity.stack.cloudvps.com/v2.0',
    'user': 'beeldbank',
    'key': os.getenv('BEELDBANK_OBJECTSTORE_PASSWORD', 'insecure'),
    'tenant_name': 'BGE000081_Cultuur',
    'os_options': {
        'tenant_id': '2b4fd2c3fcc544d2a2f1c4256098e84d',
        'region_name': 'NL',
        # 'endpoint_type': 'internalURL'
    }
}

container = 'beeldbank'
import_folder = 'Import'
images_folder = 'Images'
download_dir = '/app/data/'


@lru_cache(maxsize=None)
def get_conn():
    assert os.getenv('CULTUUR_OBJECTSTORE_PASSWORD')
    return Connection(**os_connect)


def get_full_container_list(container_name, **kwargs):
    """
    Return a listing of filenames in container `container_name`
    :param container_name:
    :param kwargs:
    :return:
    """
    limit = 10000
    kwargs['limit'] = limit
    seed = []
    _, page = get_conn().get_container(container_name, **kwargs)
    seed.extend(page)

    while len(page) == limit:
        # keep getting pages..
        kwargs['marker'] = seed[-1]['name']
        _, page = get_conn().get_container(container_name, **kwargs)
        seed.extend(page)
    return seed


def split_prefix(lst):
    """
    splits of all but the last
    """
    return '_'.join(lst.split('_')[:-1])


def copy_file_from_objectstore(file_name):
    os.makedirs(download_dir + import_folder, exist_ok=True)
    destination = download_dir + file_name
    log.info("Download file %s to %s", file_name, destination)
    with open(destination, 'wb') as f:
        f.write(get_conn().get_object(container, file_name)[1])
    return destination


def fetch_import_file_names():
    files = []

    for file_object in get_full_container_list(
            container, prefix=import_folder):
        if file_object['content_type'] == 'application/directory':
            continue

        log.info("Found file %s", file_object['name'])
        files.append(file_object['name'])
    return files
