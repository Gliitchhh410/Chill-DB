#!/bin/bash

COMMAND=$1
DB_NAME=$2
TABLE_NAME=$3


DB_DIR="./data/$DB_NAME"
TABLE_FILE="$DB_DIR/$TABLE_NAME.csv" 
META_FILE="$DB_DIR/$TABLE_NAME.meta"


if [ ! -d "data" ]; then
    echo "Error: Main data directory not found, create a DB first"
    exit 1
fi

    case $COMMAND in 
            "create")

            shift 3
            COLS_ARRAY=("$@")

            if [[ -z "$DB_NAME" || -z "$TABLE_NAME" || ${#COLS_ARRAY[@]} -eq 0 ]]; then
                echo "Error: Usage: ./table_ops.sh create <db_name> <table_name> <col:type> ..."
                exit 1
            fi

            if [ -f "$TABLE_FILE" ]; then
                echo "Error: Table '$TABLE_NAME' already exists"
                exit 1
            fi


            touch "$TABLE_FILE"
            IFS=,

            echo "${COLS_ARRAY[*]}" > "$META_FILE"
            
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

            ls "./data/$DB_NAME" | grep "\.csv$" | sed 's/\.csv$//'
            exit 0
            ;;
        "drop")
            if [[ -z "$DB_NAME" || -z "$TABLE_NAME" ]]; then
                echo "Error: Usage: ./table_ops.sh drop <db_name> <table_name>"
                exit 1
            fi

            if [ ! -f "./data/$DB_NAME/$TABLE_NAME.csv" ]; then
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


