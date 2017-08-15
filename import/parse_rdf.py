
import logging
import rdflib
import glob

from multiprocessing import Pool

from rdflib.store import Store

from rdflib_sqlalchemy import registerplugins

registerplugins()


uri = rdflib.Literal("postgresql+psycopg2://beeldbank:insecure@127.0.0.1:5432/beeldbank")  # noqa

store = rdflib.plugin.get("SQLAlchemy", Store)(identifier='test')

graph = rdflib.Graph(store, identifier='test')

graph.open(uri, create=True)
# destroy exisitng tables
graph.destroy(store)
store.create_all()

# remove existing stuff
# graph.remove()

log = logging.getLogger(__name__)
logging.basicConfig(level=logging.DEBUG)


def find_xml_files():

    all_xml_files = glob.glob('data/*.xml.clean')

    for rdf_file in all_xml_files:
        log.debug('Found:  "%s"', rdf_file)

    return all_xml_files


def load_single_file(cleaned_file):
    """
    For each xml file we craete a sperate store
    """

    log.debug('Loading %s:', cleaned_file)
    store = rdflib.plugin.get("SQLAlchemy", Store)(identifier=cleaned_file)
    graph = rdflib.Graph(store, identifier=cleaned_file)
    graph.open(uri, create=True)
    graph.destroy(store)
    store.create_all()

    graph.parse(
        cleaned_file,
        format="application/rdf+xml"
    )

    graph_stats(cleaned_file, graph)


def parse_xml_files(files):

    with Pool(3) as pool:

        pool.map(load_single_file, files)


def graph_stats(filename, bb_graph):
    """
    beeldbank graph
    """
    log.debug('%s  %s', filename, len(bb_graph))


def run_import():
    """
    Load all rdf beelbank files
    """

    files = find_xml_files()
    parse_xml_files(files)


if __name__ == '__main__':
    run_import()
