"""
WIP merge adres indicators to BAG locations

sql recrpes..

/*
  * Copy usefull bag locations into seperate table.
  * We are not interested in "toevoegingen"
  */

SELECT n."_openbare_ruimte_naam", huisnummer, geometrie
INTO bag_locaties FROM bag_nummeraanduiding n, bag_verblijfsobject v
WHERE n.verblijfsobject_id = v.id;


select count(*) from bag_locaties;

create index on bag_locaties (_openbare_ruimte_naam);

create index on bag_locaties (huisnummer);


ALTER TABLE bag_locaties ADD column id serial PRIMARY KEY;

/* DELETE duplicate entries ( we do not need H 1,2,3 *) AKA Toevoeginen */
DELETE FROM bag_locaties b1
USING bag_locaties b2
WHERE b1."_openbare_ruimte_naam"=b2."_openbare_ruimte_naam"
AND b1.huisnummer = b2.huisnummer
AND b1.id<b2.id

/* order by (b."_openbare_ruimte_naam", b.huisnummer) */

select count(*) from bag_locaties

/*
 *  Match all image locations with streetname and even / uneven street numbers]
 *
 */
SELECT
    i.id,
    image_id,
    number_from, number_to,
    streetname, b.huisnummer,
    b.id as bag_locatie_id, b.geometrie
INTO image_bag_locaties
FROM image_locations i, bag_locaties b
WHERE i.streetname = b."_openbare_ruimte_naam"
AND b.huisnummer IN (
    SELECT *
    FROM generate_series(i.number_from, i.number_to, 2))

/*
 * Match all image locations which do not have a match with
 * a broader range of numbers
 * in case of destruction of now non existing numbers,
 * something might be found nearby
 *
 */
ALTER TABLE image_bag_locaties ADD COLUMN matchtype int

CREATE unique index tada ON image_bag_locaties (id, image_id, bag_locatie_id)

CREATE index ON image_bag_locaties (image_id)

/*
 *
 *
 */
SELECT
    i.id,
    i.image_id,
    i.number_from,
    i.number_to,
    i.streetname,
    b.huisnummer,
    b.id AS bag_locatie_id, b.geometrie
INTO image_bag_locaties_2
FROM image_locations i, bag_locaties b
WHERE i.streetname = b."_openbare_ruimte_naam"
AND b.huisnummer IN (
    select * from generate_series(i.number_from -10, i.number_to + 10, 1))
AND NOT EXISTS (
    SELECT image_id
    FROM image_bag_locaties bi where bi.image_id=i.image_id)




SELECT
    image_id, streetname, number_from, number_to,
    huisnummer, bag_locatie_id, geometrie
FROM image_bag_locaties;

create index on image_bag_locaties (image_id)

select count(*), image_id
from image_bag_locaties i1
group by image_id

select count(distinct image_id) from image_bag_locaties;
select count(distinct image_id) from image_bag_locaties_2;


select count(*) from beeldbank_images;

select count(*), image_id, b.id,
from image_bag_locaties group by id, image_id, bag_locatie_id;


create index on image_bag_locaties_2 (image_id)

select count(*) from image_bag_locaties bi1, image_bag_locaties_2 bi2
where bi1.image_id!=bi2.image_id

select ib.bag_locatie_id, ib.geometrie, i.image_id, i.date_from::int
into geo_plaatjes_2
from image_bag_locaties_2 ib
left join beeldbank_images i on (ib.image_id=i.image_id);

create index on geo_plaatjes_2 (geometrie);
create index on geo_plaatjes_2 (date_from);
create index on geo_plaatjes_2 (image_id);

ALTER table geo_plaatjes_2 ADD column id serial PRIMARY KEY;


select ib.bag_locatie_id, ib.geometrie, i.image_id, i.date_from::int
into geo_plaatjes_1
from image_bag_locaties ib
left join beeldbank_images i on (ib.image_id=i.image_id);


create index on geo_plaatjes_1 (geometrie);
create index on geo_plaatjes_1 (date_from);
create index on geo_plaatjes_1 (image_id);

ALTER table geo_plaatjes_1 ADD column id serial PRIMARY KEY;

"""
