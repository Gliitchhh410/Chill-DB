#!/bin/bash


COMMAND=$1
DB_NAME=$2
TABLE_NAME=$3


if [ ! -d "./data" ]; then
    echo "Error: Main data directory not found."
    exit 1
fi

TABLE_FILE="./data/$DB_NAME/$TABLE_NAME.csv"
META_FILE="./data/$DB_NAME/$TABLE_NAME.csv"


if [[ ! -d "./data/$DB_NAME" ]]; then
    echo "Error: Database '$DB_NAME' does not exist."
    exit 1
fi

if [[ ! -d "$TABLE_FILE" ]]; then
    echo "Error: Database '$TABLE_NAME' does not exist."
    exit 1
fi


case $COMMAND in 
        "insert")
            VALUES=$4
             if [ -z "$VALUES" ]; then
                echo "Error: Usage: ./data_ops.sh insert <db_name> <table_name> <values>"
                exit 1
            fi

            PK_VALUE=$(echo "$VALUES" | cut -d',' -f1)


            if grep -q "^$PK_VALUE," "$TABLE_FILE"; then
                echo "Error: Primary Key '$PK_VALUE' already exists."
                exit 1
            fi

            echo "$VALUES" >> "$TABLE_FILE"
            echo "Row inserted successfully"
            exit 0
            ;; 


        "select")
            PK_VALUE=$4

            if [ -z "$PK_VALUE" ]; then
                cat "$TABLE_FILE"
            else 
                RESULT=$(grep "^$PK_VALUE," "$TABLE_FILE")
                if [ -z "$RESULT" ]; then
                    echo "Error: Record not found"
                    exit 1
                else
                    echo "$RESULT"
                fi
            fi
            exit 0
            ;;
            
        *)
            echo "Error: Unknown command $COMMAND"
            exit 1
            ;;
esac