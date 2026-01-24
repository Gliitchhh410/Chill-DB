#!/bin/bash


COMMAND=$1
DB_NAME=$2
TABLE_NAME=$3
VALUES=$4


if [ ! -d "./data" ]; then
    echo "Error: Main data directory not found."
    exit 1
fi


case $COMMAND in 
        "insert")
             if [[ -z "$DB_NAME" || -z "$TABLE_NAME" || -z "$VALUES" ]]; then
                echo "Error: Usage: ./data_ops.sh insert <db_name> <table_name> <values>"
                exit 1
            fi

            if [ ! -f "./data/$DB_NAME/$TABLE_NAME.csv" ]; then
                echo "Error: Table $TABLE_NAME does not exist."
                exit 1
            fi

            PK_VALUE=$(echo "$VALUES" | cut -d',' -f1)


            if grep -q "^$PK_VALUE," "./data/$DB_NAME/$TABLE_NAME.csv"; then
                echo "Error: Primary Key '$PK_VALUE' already exists."
                exit 1
            fi

            echo "$VALUES">>"./data/$DB_NAME/$TABLE_NAME.csv"

            echo "Row inserted successfully"
            exit 0
            ;; 
        "select")
            cat "./data/$DB_NAME/$TABLE_NAME.csv"
            exit 0
            ;;
        *)
            echo "Error: Unknown command $COMMAND"
            exit 1
            ;;
esac