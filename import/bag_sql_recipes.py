"""
WIP merge adres indicators to BAG locations

sql recipe
"""

import os
import psycopg2
import logging

logging.basicConfig(level=logging.DEBUG)
log = logging.getLogger(__file__)

Config = dict(
    DEBUG=os.getenv('BEELDBANK_DEBUG', True),
    PORT=os.getenv('BEELDBANK_PORT', 5432),
    HOST=os.getenv('BEELDBANK_HOST', "database"),
    USER=os.getenv('BEELDBANK_USER', "beeldbank"),
    DATABASE=os.getenv('BEELDBANK_DATABASE', "beeldbank"),
    PASSWORD=os.getenv('BEELDBANK_PASSWORD', "insecure"),
)

# global connection
connection = {}


def get_db_connection():
    c = Config
    connection = psycopg2.connect(f"""
        dbname={c['DATABASE']}
        user={c['USER']}
        host={c['HOST']}
        password={c['PASSWORD']}
        port={c['PORT']}
    """)
    return connection


def drop_table(table_name):

    with connection.cursor() as cursor:
        cursor.execute(f"DROP TABLE IF EXISTS {table_name}")


def drop_view(table_name):

    with connection.cursor() as cursor:
        cursor.execute(f"DROP VIEW IF EXISTS {table_name}")


def count_table(table_name):
    with connection.cursor() as cursor:
        cursor.execute(f"SELECT COUNT(*) from {table_name}")
        count = cursor.fetchone()[0]
        log.debug(count)
        return count


def execute_sql(sql):
    with connection.cursor() as cursor:
        cursor.execute(sql)


def create_index(table, column):
    with connection.cursor() as cursor:
        cursor.execute(f"""
CREATE INDEX ON {table} ({column})
        """)


def create_bag_table():
    """
    Given nummeraanduidingen and verblijfsobjecten
    create a table of bag locations which contain
    point locations without double entries

    We are NOT interested in 'toevoegingen'
    """
    drop_table('bag_locaties')

    log.debug('Select bag locations usefull for beeldank')
    bag_selection_sql = """

SELECT n."_openbare_ruimte_naam", huisnummer, geometrie
INTO bag_locaties
FROM bag_nummeraanduiding n, bag_verblijfsobject v
WHERE n.verblijfsobject_id = v.id;
    """

    execute_sql(bag_selection_sql)

    # make sure we have data..
    count1 = count_table('bag_locaties')
    assert count1 > 0

    # create some indexes
    create_indexes = """
    create index on bag_locaties (_openbare_ruimte_naam);
    create index on bag_locaties (huisnummer);
    ALTER TABLE bag_locaties ADD column id serial PRIMARY KEY;
    """
    execute_sql(create_indexes)

    # DELETE duplicate entries ( we do not need H 1,2,3 *) AKA Toevoeginen */
    delete_dupes_sql = """
DELETE FROM bag_locaties b1
USING bag_locaties b2
WHERE b1."_openbare_ruimte_naam"=b2."_openbare_ruimte_naam"
AND b1.huisnummer = b2.huisnummer
AND b1.id<b2.id
AND b2.geometrie is not null
    """

    execute_sql(delete_dupes_sql)
    count2 = count_table('bag_locaties')
    # we must have a lot less locations now
    assert count2 < count1

    log.info('Bag locaties prepared: %s', count2)


def simple_match_imagelocations():
    """
    Match all image locations with streetname and even /
    uneven street numbers.
    """
    im1_bag = 'image_bag_locaties_1'
    drop_table(im1_bag)
    log.debug('Match bag locations within location bounds.')

    match_sql = f"""
SELECT
    i.id,
    image_id,
    number_from, number_to,
    streetname, b.huisnummer,
    b.id as bag_locatie_id, b.geometrie
INTO {im1_bag}
FROM image_locations i, bag_locaties b
WHERE i.streetname = b."_openbare_ruimte_naam"
AND b.huisnummer IN (
    SELECT *
    FROM generate_series(i.number_from, i.number_to, 2))
    """

    execute_sql(match_sql)
    count = count_table(im1_bag)
    assert count > 100000
    log.debug('Image bag locaties: %s', count)

    validate_sql = f"""
CREATE unique index ON {im1_bag} (id, image_id, bag_locatie_id)
    """
    execute_sql(validate_sql)

    # create a needed index.
    create_index(im1_bag, 'image_id')
    create_index(im1_bag, 'streetname')


def wider_match_range_imagelocations():
    """
    Match all image locations which do not have a match with
    a broader range of numbers
    in case of destruction of now non existing numbers,
    something might be found nearby
    """
    im2_table = 'image_bag_locaties_2'
    drop_table(im2_table)

    log.debug('Match image wider locations -15, +15')
    match_wider_sql = f"""
SELECT
    i.id,
    i.image_id,
    i.number_from,
    i.number_to,
    i.streetname,
    b.huisnummer,
    b.id AS bag_locatie_id, b.geometrie
INTO {im2_table}
FROM image_locations i, bag_locaties b
WHERE i.streetname = b."_openbare_ruimte_naam"
AND b.huisnummer IN (
    select * from generate_series(i.number_from -15, i.number_to + 15, 1))
AND NOT EXISTS (
    SELECT image_id
    FROM image_bag_locaties_1 bi where bi.image_id=i.image_id)
    """
    execute_sql(match_wider_sql)

    # create a needed index.
    create_index(im2_table, 'image_id')
    create_index(im2_table, 'streetname')

    validate_sql = f"""
CREATE unique index ON {im2_table} (id, image_id, bag_locatie_id)
    """
    execute_sql(validate_sql)

    count = count_table(im2_table)
    assert count > 10000


def image_stats():
    # check results & check counts
    count1 = check_distinct_image_counts('image_bag_locaties_1')
    count2 = check_distinct_image_counts('image_bag_locaties_2')

    total_images_locations = check_distinct_image_counts('image_locations')

    log.debug(
        'Bag1 %s, Bag2 %s total %s locations %.2f %%.',
        count1, count2, total_images_locations,
        (count1 + count2) / total_images_locations * 100)


def check_distinct_image_counts(table):

    with connection.cursor() as cursor:

        cursor.execute(f"""
SELECT count(distinct image_id) FROM {table};
        """)
        image_count = cursor.fetchone()[0]
        return image_count


def validate_image_tables():
    """
    Make sure there are no images in image bag 2 (wider match)
    that are ALSO in the image_bag_locaties_1 table (simple match)
    """
    log.debug('Check for invalid locations')

    validate_sql = """
select count(*) from image_bag_locaties_1 bi1, image_bag_locaties_2 bi2
where bi1.image_id = bi2.image_id
    """
    with connection.cursor() as cursor:
        cursor.execute(validate_sql)
        count = cursor.fetchone()[0]
        log.debug('checking if we have 1 type of location for images')
        assert count == 0


def create_geo_tables(source, geo_name):
    """
    Create geo tables with `date_from` and `geometrie`
    """
    log.debug('Create view %s from %s', geo_name, source)
    drop_view(geo_name)

    execute_sql(f"""
SELECT
    ib.bag_locatie_id,
    ib.geometrie,
    i.image_id,
    i.date_from::int
INTO {geo_name}
FROM {source} ib
LEFT JOIN beeldbank_images i on (ib.image_id=i.image_id);

CREATE INDEX ON {geo_name} (geometrie);
CREATE index ON {geo_name} (date_from);
CREATE index ON {geo_name} (image_id);

ALTER table {geo_name} ADD column id serial PRIMARY KEY;
""")

    count = count_table(geo_name)
    log.debug('%s has count %s', geo_name, count)
    assert count > 1000

"""
select ib.bag_locatie_id, ib.geometrie, i.image_id, i.date_from::int
into geo_plaatjes_1
from image_bag_locaties ib
left join beeldbank_images i on (ib.image_id=i.image_id);


create index on geo_plaatjes_1 (geometrie);
create index on geo_plaatjes_1 (date_from);
create index on geo_plaatjes_1 (image_id);

ALTER table geo_plaatjes_1 ADD column id serial PRIMARY KEY;

"""

if __name__ == '__main__':
    connection = get_db_connection()

    # create_bag_table()
    simple_match_imagelocations()
    wider_match_range_imagelocations()
    validate_image_tables()
    image_stats()
    create_geo_tables('image_bag_locaties_1', 'geo_image_locaties_1')
    create_geo_tables('image_bag_locaties_2', 'geo_image_locaties_2')
