#!/bin/bash

set -e
set -u

# download data
python objectstore.py

xmlparser

python bag_sql_recipes.py
