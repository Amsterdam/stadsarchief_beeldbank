#!/bin/bash


set -e
set -u


for xml in data/*.xml; do
#for xml in 'data/b2_02_20170713_121014.xml'; do
    #tail -n 200 $xml |
    echo $xml
    grep -v 'name="geografische naam"' $xml | \
    grep -v 'name="geografische naam"' | \
    grep -v 'name="documenttype"' | \
    grep -v 'name="collectie"' | \
    grep -v 'name="rechthebbende"' | \
    grep -v 'name="download of print"' | \
    sed 's/^<sk:parameter name="datering">\(.*\)<\/sk:parameter>$/<sk:date>\1<\/sk:date>/g' |
    sed 's/^<sr:parameter name="levering">\(.*\)<\/sr:parameter>$/<sr:levering>\1<\/sr:levering>/g' |
    sed 's/^<sr:parameter name="leveringsvoorwaarden">\(.*\)<\/sr:parameter>/<sr:leveringsvoorwaarden>\1<\/sr:leveringsvoorwaarden>/g' > \
    "$xml.clean"
    echo "$xml.clean"
done

#
#    grep -v GRANT | \
#    grep -v TRIGGER | \
#    grep -v DROP | \
#    grep -v "CREATE INDEX" | \
#    grep -v "ADD CONSTRAINT" | \
#    grep -v "ALTER TABLE" | \
#    grep -v "PRIMARY" | \
#    sed 's/^.*geometry(Point.*$/    geopunt GEOMETRY(Point,28992)/' | \
#    sed 's/igp_sw44z0001_cmg_owner\.//' | \
#    psql -v ON_ERROR_STOP=1 -d handelsregister -h database -U handelsregister
# done


