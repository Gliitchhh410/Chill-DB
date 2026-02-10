#!/bin/bash

PS3="DBMS >> " # Custom prompt string

print_header() {
    clear
    echo "======================================"
    echo "      Bash DBMS - Main Menu"
    echo "======================================"
}

# ----------------------------
# The Table Menu (Inner Loop)
# ----------------------------
table_menu() {
    DB_NAME=$1
    echo "Connected to Database: $DB_NAME"
    
    select choice in "Create Table" "List Tables" "Drop Table" "Insert Row" "Select Row" "Delete Row" "Update Row" "Back to Main Menu"; do
        case $choice in
            "Create Table")
                read -p "Enter Table Name: " TABLE
                read -p "Enter Columns (e.g. id:int,name:string): " COLS
                ./table_ops.sh create "$DB_NAME" "$TABLE" "$COLS"
                ;;
            "List Tables")
                ./table_ops.sh list "$DB_NAME"
                ;;
            "Drop Table")
                read -p "Enter Table Name: " TABLE
                ./table_ops.sh drop "$DB_NAME" "$TABLE"
                ;;
            "Insert Row")
                read -p "Enter Table Name: " TABLE
                read -p "Enter Values (e.g. 1,John): " VALS
                ./data_ops.sh insert "$DB_NAME" "$TABLE" "$VALS"
                ;;
            "Select Row")
                read -p "Enter Table Name: " TABLE
                read -p "Enter PK Value (Leave empty for all): " PK
                ./data_ops.sh select "$DB_NAME" "$TABLE" "$PK"
                ;;
            "Delete Row")
                read -p "Enter Table Name: " TABLE
                read -p "Enter PK Value: " PK
                ./data_ops.sh delete "$DB_NAME" "$TABLE" "$PK"
                ;;
            "Update Row")
                read -p "Enter Table Name: " TABLE
                read -p "Enter PK Value: " PK
                read -p "Enter Column Name to change: " COL
                read -p "Enter New Value: " VAL
                ./data_ops.sh update "$DB_NAME" "$TABLE" "$PK" "$COL" "$VAL"
                ;;
            "Back to Main Menu")
                break
                ;;
            *) 
                echo "Invalid option." 
                ;;
        esac
        echo ""
        read -p "Press Enter to continue..."
    done
}

# ----------------------------
# The Main Menu (Outer Loop)
# ----------------------------
while true; do
    print_header
    select choice in "Create Database" "List Databases" "Connect To Database" "Drop Database" "Exit"; do
        case $choice in
            "Create Database")
                read -p "Enter DB Name: " DBNAME
                ./db_ops.sh create "$DBNAME"
                ;;
            "List Databases")
                ./db_ops.sh list
                ;;
            "Connect To Database")
                read -p "Enter DB Name to Connect: " DBNAME
                if [ -d "../data/$DBNAME" ]; then
                    table_menu "$DBNAME"
                else
                    echo "Error: Database '$DBNAME' not found."
                fi
                ;;
            "Drop Database")
                read -p "Enter DB Name to Drop: " DBNAME
                ./db_ops.sh drop "$DBNAME"
                ;;
            "Exit")
                echo "Goodbye!"
                exit 0
                ;;
            *) 
                echo "Invalid option." 
                ;;
        esac
        echo ""
        read -p "Press Enter to continue..."
        break 
    done
done