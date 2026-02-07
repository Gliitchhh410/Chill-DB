const dbListContainer = document.getElementById("db-list")
const sqlInput = document.getElementById('sql-input');
let activeDB = null;
let currentTable = null;
let currentDB = null; 

const Modal = {
    overlay: document.getElementById('modal-overlay'),
    title: document.getElementById('modal-title'),
    message: document.getElementById('modal-message'),
    input: document.getElementById('modal-input'),
    btnConfirm: document.getElementById('modal-confirm'),
    btnCancel: document.getElementById('modal-cancel'),

    open: function({ title, msg, showInput = false, onConfirm }) {
        this.title.textContent = title;
        this.message.textContent = msg;
        
        if (showInput) {
            this.input.classList.remove('hidden');
            this.input.value = ''; 
            this.input.focus();   
        } else {
            this.input.classList.add('hidden');
        }

        this.overlay.classList.remove('hidden');

        this.btnCancel.onclick = () => {
            this.close();
        };

        this.btnConfirm.onclick = () => {
            const inputValue = this.input.value;

            if (showInput && !inputValue.trim()) return;
            
            this.close();
            if (onConfirm) onConfirm(inputValue); 
        };
    },

    close: function() {
        this.overlay.classList.add('hidden');
    }
};

async function fetchDatabases(){
    try {
        const res = await fetch("/databases")

        if (!res.ok){
            throw new Error("Network response is not ok!!")
        }

        const text = await res.text()
        const databases = text.split('\n').filter(line => line.trim() !=='')

        dbListContainer.innerHTML = ''

        databases.forEach(dbName => {
            const item = document.createElement('div');
            item.className = 'p-3 hover:bg-gray-700 cursor-pointer rounded text-gray-300 text-sm font-medium transition-colors mb-1 flex justify-between items-center group';
            
            item.innerHTML = `
                <span onclick="fetchTables('${dbName}')" class="flex-1 flex items-center">
                    📂 <span class="ml-2">${dbName.slice(0,-1)}</span>
                </span>
                
                <button onclick="dropDatabase('${dbName}')" class="text-gray-500 hover:text-red-500 opacity-0 group-hover:opacity-100 transition px-2">
                    ✕
                </button>
            `;
            
            dbListContainer.appendChild(item);
        });
    } catch (e){
        console.error(`Error fetching the databases: ${e}`);
        dbListContainer.innerHTML = '<div class="text-red-500 p-2 text-xs">Failed to load</div>'
    }
}

async function fetchTables(dbName){
    currentDB = dbName
    const mainView = document.getElementById("main-view")
    mainView.innerHTML=`
        <div class="flex flex-col items-center justify-center h-full text-gray-500">
            <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-500 mb-4"></div>
            <p>Loading tables from ${dbName}...</p>
        </div>
    `;

    try {
        const res = await fetch("/tables", {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({name: dbName})
        })

        if (!res.ok) throw new Error("Failed to load tables")

        const text = await res.text()
        const tables = text.split('\n').filter(line => line.trim() !== '');

        mainView.innerHTML = '';
        const header = document.createElement('div');
        header.className = 'mb-6 flex justify-between items-center';
        header.innerHTML = `
            <h2 class="text-2xl font-bold text-white flex items-center">
                <span class="text-blue-400 mr-2">📂</span> ${dbName.slice(0,-1) }
            </h2>
            <button onclick="createTable('${dbName}')" class="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded text-sm font-medium transition">
                + New Table
            </button>
        `;
        mainView.appendChild(header);

        const grid = document.createElement('div');
        grid.className = 'grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4';

        if (tables.length === 0) {
            grid.innerHTML = '<p class="text-gray-500 col-span-3 italic">No tables found. Create one!</p>';
        }

        tables.forEach(tableName => {
            const card = document.createElement('div');
            card.className = 'bg-gray-800 border border-gray-700 p-4 rounded-lg hover:border-blue-500 transition cursor-pointer group relative';

            card.innerHTML = `
                <div class="flex items-center justify-between">
                    <div class="flex items-center">
                        <span class="text-2xl mr-3">📄</span>
                        <div>
                            <h3 class="font-bold text-gray-200">${tableName}</h3>
                            <p class="text-xs text-gray-500">Table</p>
                        </div>
                    </div>
                    <button onclick="event.stopPropagation(); dropTable('${dbName}', '${tableName}')" 
                            class="text-red-500 opacity-0 group-hover:opacity-100 hover:text-red-400 transition p-1">
                        🗑️
                    </button>
                </div>
            `;

            card.onclick = () => fetchTableData(dbName, tableName);

            grid.appendChild(card);
        });
        

        mainView.appendChild(grid);
    } catch (error) {
        console.error(error);
        mainView.innerHTML = `<div class="text-red-500">Error loading tables: ${error.message}</div>`;
    }
}


async function fetchTableData(dbName, tableName) {
    const mainView = document.getElementById('main-view');
    activeDB = dbName;
    currentTable = tableName;
    mainView.innerHTML = `
        <div class="flex flex-col items-center justify-center h-full text-gray-400">
            <div class="animate-pulse mb-4 text-4xl">📄</div>
            <p>Reading ${tableName}...</p>
        </div>
    `;

    try {
        const response = await fetch('/data/query', { 
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ 
                db_name: dbName, 
                table_name: tableName,
                column: "", 
                value: "" 
            })
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || "Failed to fetch data");
        }

        const data = await response.json(); 
        renderDataGrid(dbName, tableName, data.columns, data.rows);

    } catch (error) {
        console.error(error);
        mainView.innerHTML = `<div class="text-red-500 p-4">Error loading data: ${error.message}</div>`;
    }
}

function renderDataGrid(dbName, tableName, columns, rows) {
    const mainView = document.getElementById('main-view');
    mainView.innerHTML = '';

    const header = document.createElement('div');
    header.className = 'flex flex-col md:flex-row items-start md:items-center justify-between mb-6 gap-4 px-1 md:px-0';
    
    header.innerHTML = `
        <div class="flex items-center w-full md:w-auto">
            <button onclick="fetchTables('${dbName}')" class="text-gray-400 hover:text-white mr-3 md:mr-4 transition flex items-center shrink-0">
                <span class="text-lg md:text-xl mr-1">←</span> <span class="hidden md:inline">Back</span>
            </button>
            
            <h2 class="text-xl md:text-3xl font-bold text-gray-100 flex items-center tracking-tight truncate">
                <span class="text-blue-500 mr-2 md:mr-3 text-lg md:text-2xl">📄</span> ${tableName}
            </h2>
            
            <span class="ml-3 md:ml-4 text-[10px] md:text-xs font-mono text-blue-300 bg-blue-900 bg-opacity-30 px-2 py-1 rounded border border-blue-800 shrink-0">
                ${rows.length} records
            </span>
        </div>
        
        <button onclick="handleInsert('${dbName}', '${tableName}')"
                class="bg-blue-600 hover:bg-blue-500 text-white rounded-lg shadow-lg transition-transform transform hover:scale-105 active:scale-95 font-medium
                        w-full py-3 text-sm        md:w-auto md:px-5 md:py-2   ">
            + Add Row
        </button>
    `; 
    mainView.appendChild(header);

    const tableContainer = document.createElement('div');
    tableContainer.className = 'bg-gray-800 rounded-xl shadow-2xl border border-gray-700 overflow-hidden';

    let html = `
        <div class="overflow-x-auto">
            <table class="w-full text-left text-sm text-gray-300">
                <thead class="bg-gray-900 text-gray-400 uppercase font-semibold tracking-wider">
                    <tr>
                        ${columns.map(colName => `<th class="px-4 py-3 md:px-6 md:py-4 border-b border-gray-700 whitespace-nowrap">${colName}</th>`).join('')}
                        <th class="px-4 py-3 md:px-6 md:py-4 border-b border-gray-700 text-right whitespace-nowrap">Actions</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-700">
    `;

    rows.forEach(row => {
        const pkVal = row[0]; 
        const pkCol = columns[0]; 
        
        html += `<tr class="hover:bg-gray-700 transition-colors duration-150 group">`;
        
        row.forEach(cell => {
            html += `<td class="px-4 py-3 md:px-6 md:py-4 whitespace-nowrap group-hover:text-white">${cell}</td>`;
        });

        html += `
            <td class="px-4 py-3 md:px-6 md:py-4 text-right whitespace-nowrap">
                <button onclick="deleteRow('${dbName}', '${tableName}', '${pkCol}', '${pkVal}')"
                        class="text-red-500 hover:text-red-400 transition transform hover:scale-110 p-2
                               opacity-100 md:opacity-0 md:group-hover:opacity-100"
                        title="Delete Row">
                    🗑️
                </button>
            </td>
        </tr>`;
    });

    html += `</tbody></table></div>`;
    
    if(rows.length === 0) {
        html += `<div class="p-8 text-center text-gray-500">No records found. Click "Add Row" to start.</div>`
    }

    tableContainer.innerHTML = html;
    mainView.appendChild(tableContainer);
}

function handleInsert(dbName, tableName) {
    Modal.open({
        title: 'Insert Row',
        msg: `Enter values for ${tableName} (separated by comma):`,
        showInput: true,
        onConfirm: async (values) => {
            if (!values) return;

            try {
                const response = await fetch('/data/insert', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        db_name: dbName,
                        table_name: tableName,
                        values: values
                    })
                });

                if (response.ok) {
                    console.log("Insert successful");
                    await fetchTableData(dbName, tableName); 
                } else {
                    const result = await response.text();
                    Modal.open({ title: 'Insert Failed', msg: "Error: " + result, onConfirm: () => {} });
                }

            } catch (error) {
                console.error("Insert failed:", error);
                Modal.open({ title: 'Connection Error', msg: "Failed to connect to server.", onConfirm: () => {} });
            }
        }
    });
}


function deleteRow(dbName, tableName, colName, colValue) {
    Modal.open({
        title: 'Delete Row?',
        msg: `Are you sure you want to permanently delete the record where ${colName} = "${colValue}"? This action cannot be undone.`,
        showInput: false, 
        onConfirm: async () => {
            try {
                const query = `DELETE FROM ${tableName} WHERE ${colName}=${colValue}`;

                const response = await fetch('/sql', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        db_name: dbName,
                        query: query 
                    })
                });
                
                if (response.ok) {
                    console.log("Delete successful");
                    fetchTableData(dbName, tableName);
                } else {
                    const err = await response.text();
                    Modal.open({ title: 'Error', msg: "Error: " + err, onConfirm: () => {} }); 
                }
            } catch (error) {
                console.error(error);
                Modal.open({ title: 'System Error', msg: error.message, onConfirm: () => {} });
            }
        }
    });
}


function promptInsertRow(dbName, tableName) {
    Modal.open({
        title: 'Add New Row',
        msg: `Enter values for ${tableName} (comma separated, e.g., "5,Sarah"):`,
        showInput: true, 
        onConfirm: async (inputValue) => {
            try {
                const response = await fetch('/data/insert', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        db_name: dbName,
                        table_name: tableName,
                        values: inputValue
                    })
                });

                if (response.ok) {
                    fetchTableData(dbName, tableName);
                } else {
                    const err = await response.text();
                    Modal.open({ title: 'Error', msg: err, onConfirm: () => {} });
                }
            } catch (error) {
                console.error(error);
                Modal.open({ title: 'Error', msg: "Failed to insert row.", onConfirm: () => {} });
            }
        }
    });
}


function createDatabase() {
    Modal.open({
        title: 'Create New Database',
        msg: 'Enter a name for your new database (no spaces):',
        showInput: true,
        onConfirm: async (dbName) => {
            if (!dbName) return;

            try {
                const response = await fetch('/database/create', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name: dbName })
                });

                const result = await response.text();

                if (response.ok) {
                    fetchDatabases();
                } else {
                    Modal.open({ title: 'Creation Failed', msg: "Error: " + result, onConfirm: () => {} });
                }
            } catch (error) {
                console.error(error);
                Modal.open({ title: 'Network Error', msg: "Failed to connect.", onConfirm: () => {} });
            }
        }
    });
}


function createTable(dbName) {
    Modal.open({
        title: 'Create New Table',
        msg: 'Step 1: Enter the Table Name:',
        showInput: true,
        onConfirm: (tableName) => {
            if (!tableName) return;

            setTimeout(() => {
                Modal.open({
                    title: 'Define Columns',
                    msg: `Step 2: Enter columns for '${tableName}' (e.g., id:int,name:string):`,
                    showInput: true,
                    onConfirm: async (columns) => {
                        if (!columns) return;
                        
                        try {
                            const response = await fetch('/table/create', {
                                method: 'POST',
                                headers: { 'Content-Type': 'application/json' },
                                body: JSON.stringify({
                                    db_name: dbName,
                                    table_name: tableName,
                                    columns: columns
                                })
                            });

                            const result = await response.text();

                            if (response.ok) {
                                fetchTables(dbName); 
                            } else {
                                Modal.open({ title: 'Table Error', msg: "Error: " + result, onConfirm: () => {} });
                            }
                        } catch (error) {
                            console.error(error);
                            Modal.open({ title: 'Connection Error', msg: "Failed to connect.", onConfirm: () => {} });
                        }
                    }
                });
            }, 300); 
        }
    });
}

function dropTable(dbName, tableName) {
    Modal.open({
        title: 'Delete Table?',
        msg: `Are you sure you want to delete table '${tableName}'? All data will be lost.`,
        onConfirm: async () => {
            try {
                const response = await fetch('/table/delete', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        db_name: dbName,
                        table_name: tableName
                    })
                });

                const result = await response.text();

                if (response.ok) {
                    fetchTables(dbName); 
                } else {
                    Modal.open({ title: 'Delete Failed', msg: "Error: " + result, onConfirm: () => {} });
                }
            } catch (error) {
                console.error(error);
                Modal.open({ title: 'Connection Error', msg: "Failed to connect.", onConfirm: () => {} });
            }
        }
    });
}

function dropDatabase(dbName) {
    Modal.open({
        title: 'Delete Database?',
        msg: `WARNING: You are about to delete '${dbName}' and ALL its tables. This cannot be undone.`,
        onConfirm: async () => {
            try {
                const response = await fetch('/database/delete', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ name: dbName })
                });

                const result = await response.text();

                if (response.ok) {
                    fetchDatabases(); 
                    document.getElementById('main-view').innerHTML = '';
                } else {
                    Modal.open({ title: 'Delete Failed', msg: "Error: " + result, onConfirm: () => {} });
                }
            } catch (error) {
                console.error(error);
                Modal.open({ title: 'Connection Error', msg: "Failed to connect.", onConfirm: () => {} });
            }
        }
    });
}

sqlInput.addEventListener("keypress", async (e)=>{
    if (e.key === "Enter"){
        const query = sqlInput.value.trim()
        if (!query) return

        sqlInput.disabled = true

        try {
            await executeSQL(query);
            
        } catch (err) {
            console.error("Critical Failure:", err);
            Modal.open({ title: 'Application Error', msg: err.message, onConfirm: () => {} });
        } finally {
            sqlInput.disabled = false;
            sqlInput.focus();
        }
    }
})

async function executeSQL(query) {
    activeDB = currentDB
    if (!activeDB) {
        Modal.open({ title: 'No Database', msg: "Please select a database first!", onConfirm: () => {} });
        return;
    }


    const sqlInput = document.getElementById('sql-input');
    if(sqlInput) sqlInput.disabled = true;

    try {
        const response = await fetch('/sql', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                db_name: activeDB,
                query: query
            })
        });

        const result = await response.text();

        if (response.ok) {
            await renderSQLResult(activeDB, result);
        } else {
            Modal.open({ title: 'SQL Error', msg: "Error: " + result, onConfirm: () => {} });
        }

    } catch (error) {
        console.error(error);
        Modal.open({ title: 'System Error', msg: error.message, onConfirm: () => {} });
    } finally {
        if(sqlInput) {
            sqlInput.disabled = false;
            sqlInput.focus();
        }
    }
}


async function renderSQLResult(dbName, csvData) {
    if (!csvData || !csvData.trim()) {
        Modal.open({ title: 'No Results', msg: "Query returned no results.", onConfirm: () => {} });
        return;
    }

    const trimmedData = csvData.trim();


    if (trimmedData.startsWith("Error") || trimmedData.startsWith("syntax error") || trimmedData.startsWith("System Error")) {
        Modal.open({ title: 'Query Error', msg: trimmedData, onConfirm: () => {} });
        return;
    }

    if (trimmedData.startsWith("Success") || trimmedData.startsWith("Table")) {
        
        if (activeDB && currentTable) {
            console.log("Action successful, refreshing table:", currentTable);
            await fetchTableData(activeDB, currentTable);
        } else {
            Modal.open({ title: 'Success', msg: trimmedData, onConfirm: () => {} });
        }
        return;
    }


    const rows = trimmedData.split('\n').map(line => line.split(','));
    if (rows.length === 0) return;

    const numberOfCols = rows[0].length;
    const columns = Array.from({ length: numberOfCols }, (_, i) => `Col ${i + 1}`);

    renderDataGrid(dbName, "SQL Result", columns, rows);
}


fetchDatabases()