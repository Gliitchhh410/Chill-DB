#!/bin/bash


COMMAND=$1
DB_NAME=$2
TABLE_NAME=$3


if [ ! -d "./data" ]; then
    echo "Error: Main data directory not found."
    exit 1
fi

TABLE_FILE="./data/$DB_NAME/$TABLE_NAME.csv"
META_FILE="./data/$DB_NAME/$TABLE_NAME.meta"


if [[ ! -d "./data/$DB_NAME" ]]; then
    echo "Error: Database '$DB_NAME' does not exist."
    exit 1
fi

if [[ ! -f "$TABLE_FILE" ]]; then
    echo "Error: Table '$TABLE_NAME' does not exist."
    exit 1
fi


case $COMMAND in 
        "insert")
            VALUES=$4

            
            if [ -z "$VALUES" ]; then
                echo "Error: Usage: ./data_ops.sh insert <db_name> <table_name> <values>"
                exit 1
            fi


            EXPECTED_COLS=$(awk -F',' '{print NF}' "$META_FILE")
            ACTUAL_COLS=$(echo "$VALUES" | awk -F',' '{print NF}')

            if [ "$EXPECTED_COLS" -ne "$ACTUAL_COLS" ]; then
            echo "Error: Column count mismatch. Table expects $EXPECTED_COLS columns, but you provided $ACTUAL_COLS."
            exit 1
            fi

            PK_VALUE=$(echo "$VALUES" | cut -d',' -f1)

            if grep -q "^$PK_VALUE\(,\|$\)" "$TABLE_FILE"; then
                echo "Error: Primary Key '$PK_VALUE' already exists."
                exit 1
            fi

            PK_VALUE=$(echo "$VALUES" | cut -d',' -f1)


            echo "$VALUES" >> "$TABLE_FILE"
                echo "Row inserted successfully"
                exit 0
            ;;


        "select")
            COL_NAME=$4
            SEARCH_VAL=$5

            if [ -z "$COL_NAME" ]; then
                cat "$TABLE_FILE"
                exit 0
            fi

            COL_NUM=$(awk -F, -v col="$COL_NAME" '{
                for(i=1;i<=NF;i++){
                    gsub(/[ \t\r\n]+$/, "", $i) 
                    split($i, def, ":");
                    if(def[1] == col) print i;
                    }
            }' "$META_FILE")

            if [ -z "$COL_NUM" ]; then
                echo "Error: Column '$COL_NAME' not found"
                exit 1
            fi

            awk -F, -v c="$COL_NUM" -v v="$SEARCH_VAL" '$c == v' "$TABLE_FILE"
            exit 0
            ;;
            
        "delete")
            PK_VALUE=$4

            if [ -z "$PK_VALUE" ]; then
                echo "Error: Usage: ./data_ops.sh delete <db> <table> <pk>"
                exit 1
            fi


            if ! grep -q "^$PK_VALUE," "$TABLE_FILE"; then
                echo "Error: Record with ID $PK_VALUE not found"
                exit 1
            fi

            awk -F, -v pk="$PK_VALUE" '$1 !=pk' "$TABLE_FILE">"$TABLE_FILE.tmp" && mv "$TABLE_FILE.tmp" "$TABLE_FILE"
            echo "Record deleted successfully."
            exit 0
            ;;
        "update")


            PK_VALUE=$4
            COL_NAME=$5
            NEW_VALUE=$6
            if [[ -z "$PK_VALUE" || -z "$COL_NAME" || -z "$NEW_VALUE" ]]; then
                echo "Error: Usage: ./data_ops.sh update <db> <table> <pk> <col> <new_val>"
                exit 1
            fi

            COL_NUM=$(awk -F, -v col="$COL_NAME" '{
                    for(i=1;i<=NF;i++){
                        split($i, def, ":");
                        if(def[1] == col) print i;
                    }
            }' "$META_FILE")

            if [ -z "$COL_NUM" ]; then
                echo "Error: Column '$COL_NAME' not found in table schema."
                exit 1
            fi

            # FIX: Prevent updating the Primary Key
            if [ "$COL_NUM" -eq 1 ]; then
                echo "Error: You cannot update the Primary Key column."
                exit 1
            fi

            if ! grep -q "^$PK_VALUE\(,\|$\)" "$TABLE_FILE"; then
                echo "Error: Record with ID '$PK_VALUE' not found."
                exit 1
            fi


            awk -F, -v pk="$PK_VALUE" -v col="$COL_NUM" -v val="$NEW_VALUE" 'BEGIN{OFS=","} {
                if ($1 == pk){
                    $col = val
                }
                print $0
            }' "$TABLE_FILE" > "$TABLE_FILE.tmp" && mv "$TABLE_FILE.tmp" "$TABLE_FILE"

            echo "Record updated successfully"
            exit 0
            ;;

        *)
            echo "Error: Unknown command $COMMAND"
            exit 1
            ;;
esac