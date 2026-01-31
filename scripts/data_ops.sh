#!/bin/bash


COMMAND=$1
DB_NAME=$2
TABLE_NAME=$3


SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
DATA_DIR="$PROJECT_ROOT/data"
DB_DIR="$DATA_DIR/$DB_NAME"
TABLE_FILE="$DB_DIR/$TABLE_NAME.csv"
META_FILE="$DB_DIR/$TABLE_NAME.meta"


if [ ! -d "$DB_DIR" ]; then
    echo "Error: Database '$DB_NAME' does not exist."
    exit 1
fi

if [ ! -f "$TABLE_FILE" ]; then
    echo "Error: Table '$TABLE_NAME' does not exist."
    exit 1
fi

if [ ! -f "$META_FILE" ]; then
    echo "Error: Meta file for '$TABLE_NAME' not found."
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
        FILTER_COL=$4
        FILTER_VAL=$5

        if [ -z "$FILTER_COL" ] || [ -z "$FILTER_VAL" ]; then
            echo "Error: Usage: ./data_ops.sh delete <db> <table> <col> <value>"
            exit 1
        fi


        HEADER=$(head -n 1 "$TABLE_FILE" | tr -d '\r')
        
        COL_IDX=$(awk -F, -v col="$FILTER_COL" '{
            for(i=1; i<=NF; i++) {
                # Split "id:int" into parts[1]="id", parts[2]="int"
                split($i, parts, ":");
                
                # Clean whitespace/newlines from the name
                col_name = parts[1];
                gsub(/^[ \t\r\n]+|[ \t\r\n]+$/, "", col_name);
                
                if(col_name == col) { 
                    print i; 
                    exit 
                }
            }
        }' "$META_FILE")

        if [ -z "$COL_IDX" ]; then
            echo "Error: Column '$FILTER_COL' not found in table header."
            exit 1
        fi


        EXISTS=$(awk -F, -v idx="$COL_IDX" -v val="$FILTER_VAL" '
            {  # Runs on ALL lines including line 1
                clean_cell=$idx; 
                gsub(/^[ \t\r\n]+|[ \t\r\n]+$/, "", clean_cell);
                if(clean_cell == val) { print "yes"; exit }
            }' "$TABLE_FILE")

        if [ "$EXISTS" != "yes" ]; then
             echo "0 rows deleted. (No match for $FILTER_COL=$FILTER_VAL)"
             exit 0
        fi


        awk -F, -v idx="$COL_IDX" -v val="$FILTER_VAL" '
            {
                clean_cell=$idx; 
                gsub(/^[ \t\r\n]+|[ \t\r\n]+$/, "", clean_cell);
                # Only print lines that DO NOT match
                if(clean_cell != val) print $0
            }' "$TABLE_FILE" > "$TABLE_FILE.tmp" && mv "$TABLE_FILE.tmp" "$TABLE_FILE"

        echo "Record(s) deleted successfully."
        exit 0
        ;;
        "update")


            DB_NAME=$2
            TABLE_NAME=$3
            SET_COL=$4
            SET_VAL=$5
            WHERE_COL=$6
            WHERE_VAL=$7
            TEMP_FILE="$DB_DIR/$TABLE_NAME.tmp"
            if [[ -z "$SET_COL" || -z "$SET_VAL" || -z "$WHERE_COL" || -z "$WHERE_VAL" ]]; then
                echo "Error: Usage: ./data_ops.sh update <db> <table> <pk> <col> <new_val>"
                exit 1
            fi

            SET_IDX=$(awk -F, -v col="$SET_COL" '{
            for(i=1;i<=NF;i++){
                split($i, def, ":");
                if(def[1] == col) {print i; exit}
            }   
            }' "$META_FILE")
    WHERE_IDX=$(awk -F, -v col="$WHERE_COL" '{
            for(i=1;i<=NF;i++){
                split($i, def, ":");
                if(def[1] == col) {print i; exit}
            }
            }' "$META_FILE")

            if [ -z "$SET_IDX" ]; then
                echo "Error: Column '$SET_COL' not found in table schema."
                exit 1
            fi

            if [ -z "$WHERE_IDX" ]; then
                echo "Error: Column '$WHERE_COL' not found"
                exit 1
            fi

            if [ "$SET_IDX" -eq 1 ]; then
                echo "Error: You cannot update the Primary Key column."
                exit 1
            fi


        awk -F, -v s_idx="$SET_IDX" -v s_val="$SET_VAL" -v w_idx="$WHERE_IDX" -v w_val="$WHERE_VAL" '
            BEGIN { OFS="," }
            {
            if ($w_idx == w_val) {
                $s_idx = s_val
            }
            print $0
            }' "$TABLE_FILE" > "$TEMP_FILE"

    mv "$TEMP_FILE" "$TABLE_FILE"
            

            exit 0
            ;;

        *)
            echo "Error: Unknown command $COMMAND"
            exit 1
            ;;
esac