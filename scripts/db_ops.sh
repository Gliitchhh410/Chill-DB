#!/bin/bash

COMMAND=$1

DB_NAME=$2

if [ ! -d "./data" ]; then
    mkdir  ./data
fi


case $COMMAND in 
        "create")
            if [ -z "$DB_NAME" ]; then
                echo "Error: Database name is required for 'create' command"
                exit 1
            fi

            if [ -d "./data/$DB_NAME" ]; then
                echo "Error: Database already exists"
                exit 1
            fi

            mkdir "./data/$DB_NAME"
            echo "$DB_NAME created successfully"
            exit 0
            ;;
        "list")
            ls ./data -F
            exit 0 
            ;;
        
        "drop")
            if [ -z "$DB_NAME" ]; then
                echo "Error: Database name is required for 'drop' command"
                exit 1
            fi

            if [ ! -d "./data/$DB_NAME" ]; then
                echo "Error: Database does not exist"
                exit 1
            fi

            rm -r "./data/$DB_NAME"

            echo "$DB_NAME deleted successfully"
            exit 0
            ;;

        *)
            echo "Error: Unknown command $COMMAND"
            echo "Usage: ./db_ops.sh {create|list|drop} [db_name]"
            exit 1
            ;;
esac
