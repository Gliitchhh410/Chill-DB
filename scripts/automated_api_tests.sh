#!/bin/bash
HOST="http://localhost:8080"
DB_NAME="TestDB"
TABLE_USERS="users"
TABLE_PRODUCTS="products"

# Colors for output
BLUE='\033[1;34m'
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}"
}

print_success() {
    echo -e "${GREEN}✔ Request Sent${NC}"
}


# PHASE 1: SETUP & BUTTON APIs


print_header "1. Create Database (Button: Create DB)"
curl -s -X POST "$HOST/database/create" \
     -H "Content-Type: application/json" \
     -d "{\"name\": \"$DB_NAME\"}"
echo ""

print_header "2. Create Tables (Button: Create Table)"
echo "Creating 'users' table..."
curl -s -X POST "$HOST/table/create" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"table_name\": \"$TABLE_USERS\", \"columns\": \"id:int,name:string,role:string\"}"
echo ""

echo "Creating 'products' table (Testing Multi-table isolation)..."
curl -s -X POST "$HOST/table/create" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"table_name\": \"$TABLE_PRODUCTS\", \"columns\": \"sku:int,item:string\"}"
echo ""

print_header "3. Inserting Data (Button: Add Row)"
echo "Inserting Alice (Admin)..."
curl -s -X POST "$HOST/data/insert" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"table_name\": \"$TABLE_USERS\", \"values\": \"1,Alice,Admin\"}"
echo ""

echo "Inserting Bob (User)..."
curl -s -X POST "$HOST/data/insert" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"table_name\": \"$TABLE_USERS\", \"values\": \"2,Bob,User\"}"
echo ""

echo "Inserting Charlie (User)..."
curl -s -X POST "$HOST/data/insert" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"table_name\": \"$TABLE_USERS\", \"values\": \"3,Charlie,User\"}"
echo ""

echo "Inserting Product (Laptop)..."
curl -s -X POST "$HOST/data/insert" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"table_name\": \"$TABLE_PRODUCTS\", \"values\": \"101,Laptop\"}"
echo ""

# PHASE 2: ERROR HANDLING (EDGE CASES)


print_header "4. Testing Edge Cases (Should Fail)"

echo "TEST: Inserting with wrong column count (4 values for 3 cols)..."
curl -s -X POST "$HOST/data/insert" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"table_name\": \"$TABLE_USERS\", \"values\": \"4,Eve,User,ExtraData\"}"
echo -e "\n(Above should be an error)"

echo "TEST: Duplicate Primary Key (Should Fail)..."
# Try to insert Alice (ID 1) again
curl -s -X POST "$HOST/data/insert" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"table_name\": \"$TABLE_USERS\", \"values\": \"1,AliceClone,Admin\"}"
echo -e "\n(Above should be an error: Primary Key exists)"

echo "TEST: Querying non-existent table..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"SELECT * FROM ghosts\"}"
echo -e "\n(Above should be an error)"


# PHASE 3: SQL ENGINE TESTS


print_header "5. SQL SELECT (The SQL Box)"

echo "Querying ALL Users (Should see 3 rows)..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"SELECT * FROM $TABLE_USERS\"}"
echo ""


echo "TEST: SELECT WHERE clause (Finding Bob)..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"SELECT * FROM $TABLE_USERS WHERE name=Bob\"}"
echo ""

echo "Querying Products (Should see 1 row, ensuring no Users leaked here)..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"SELECT * FROM $TABLE_PRODUCTS\"}"
echo ""


echo "TEST: SELECT Specific Columns (Name Only)..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"SELECT name FROM $TABLE_USERS\"}"
echo ""

print_header "6. SQL DELETE Logic (The Delete Button/SQL Box)"

echo "TEST: Delete by ID (Deleting Alice id=1)..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"DELETE FROM $TABLE_USERS WHERE id=1\"}"
echo ""

echo "TEST: Delete by String/Name (Deleting Charlie)..."
# Assuming parser handles basic string matching
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"DELETE FROM $TABLE_USERS WHERE name=Charlie\"}"
echo ""

print_header "9. Verify Results"
echo "Reading Users (Should ONLY see Bob)..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"SELECT * FROM $TABLE_USERS\"}"
echo ""

print_header "8. SQL UPDATE Logic (The Boss Level)"

echo "Current State of Bob:"
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"SELECT * FROM $TABLE_USERS WHERE name=Bob\"}"
echo ""

echo "TEST: Updating Bob's Role to 'SuperAdmin'..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"UPDATE $TABLE_USERS SET role=SuperAdmin WHERE name=Bob\"}"
echo ""

echo "VERIFY: Checking if Bob is now a SuperAdmin..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"SELECT * FROM $TABLE_USERS WHERE name=Bob\"}"
echo ""

echo "TEST: Update non-existent user (Should affect 0 rows or fail gracefully)..."
curl -s -X POST "$HOST/sql" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"query\": \"UPDATE $TABLE_USERS SET role=Ghost WHERE name=Casper\"}"
echo ""

# PHASE 4: DESTRUCTIVE CLEANUP

print_header "9. Drop Table (Button: Trash Icon)"
echo "Dropping Products table..."
curl -s -X POST "$HOST/table/delete" \
     -H "Content-Type: application/json" \
     -d "{\"db_name\": \"$DB_NAME\", \"table_name\": \"$TABLE_PRODUCTS\"}"
echo ""

echo "Verifying it's gone (Should fail)..."
curl -s -X POST "$HOST/tables" \
     -H "Content-Type: application/json" \
     -d "{\"name\": \"$DB_NAME\"}"
echo ""

print_header "10 Drop Database (Button: Delete DB)"
curl -s -X POST "$HOST/database/delete" \
     -H "Content-Type: application/json" \
     -d "{\"name\": \"$DB_NAME\"}"
echo ""





print_header "Test Suite Complete."