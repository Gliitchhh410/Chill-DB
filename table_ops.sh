#!/bin/bash

COMMAND=$1
DB_NAME=$2
TABLE_NAME=$3
COLS=$4


if [ ! -d "./data" ]; then
    echo "Error: Main data directory not found, create a DB first"
    exit 1
fi

case $COMMAND in 
        "create")
            if [[ -z "$DB_NAME" || -z "$TABLE_NAME" || -z "$COLS" ]]; then
                echo "Error: Usage: ./table_ops.sh create <db_name> <table_name> <COLS>"
                exit 1
            fi

            if [ ! -d "./data/$DB_NAME" ]; then
                echo "Error: Database $DB_NAME does not exist"
                exit 1
            fi


            if [ -f "./data/$DB_NAME/$TABLE_NAME.csv" ]; then
                echo "Error: Table $TABLE_NAME already exists"
                exit 1
            fi

            touch "./data/$DB_NAME/$TABLE_NAME.csv"

            echo "$COLS" > "./data/$DB_NAME/$TABLE_NAME.meta"
            touch "./data/$DB_NAME/$TABLE_NAME.pk"
            echo "Table '$TABLE_NAME' created successfully."
            exit 0
            ;;


        "list")

            if [[ -z "$DB_NAME" ]]; then
                echo "Error: Usage: ./table_ops.sh list <db_name>"
                exit 1
            fi

            if [ ! -d "./data/$DB_NAME" ]; then
                echo "Error: Database $DB_NAME does not exist"
                exit 1
            fi

            echo "Tables in $DB_NAME: "
            ls "./data/$DB_NAME"
            exit 0
            ;;
        "drop")
            if [[ -z "$DB_NAME" || -z "$TABLE_NAME" ]]; then
                echo "Error: Usage: ./table_ops.sh drop <db_name> <table_name>"
                exit 1
            fi

            if [ -f "./data/$DB_NAME/$TABLE_NAME.csv" ]; then
                echo "Error: Table $TABLE_NAME does not exist"
                exit 1
            fi

            rm "./data/$DB_NAME/$TABLE_NAME.csv"
            rm "./data/$DB_NAME/$TABLE_NAME.meta"
            rm "./data/$DB_NAME/$TABLE_NAME.pk"

            echo "Table $TABLE_NAME dropped successfully"
            exit 0

            ;;
        *)
            echo "Error: Uknown command $COMMAND"
            exit 1
            ;;
esac


