const dbListContainer = document.getElementById("db-list")
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
            onConfirm(inputValue); 
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


        databases.forEach(db => {
            const item = document.createElement("div")
            item.className = 'p-3 hover:bg-gray-700 cursor-pointer rounded text-gray-300 text-sm font-medium transition-colors mb-1'
            item.textContent = `📂 ${ db.slice(0,-1)}`

            item.onclick= () => {
                fetchTables(db)
            }

            dbListContainer.appendChild(item)
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
            <button onclick="promptCreateTable()" class="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded text-sm font-medium transition">
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
                    <button onclick="event.stopPropagation(); deleteTable('${dbName}', '${tableName}')" 
                            class="text-red-500 opacity-0 group-hover:opacity-100 hover:text-red-400 transition p-1">
                        🗑️
                    </button>
                </div>
            `;

            // Click to view data (Future Step)
            card.onclick = () => fetchTableData(dbName, tableName);

            grid.appendChild(card);
        });
        

        mainView.appendChild(grid);
    } catch (error) {
        console.error(error);
        mainView.innerHTML = `<div class="text-red-500">Error loading tables: ${error.message}</div>`;
    }
}


async function fetchTableData(dbName, tableName){
    const mainView = document.getElementById('main-view')

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

        const text = await response.text();
        
        if (!text || text.trim() === "") {
            renderEmptyTableState(dbName, tableName);
            return;
        }

        const rows = text.split('\n')
            .filter(row => row.trim() !== '')
            .map(row => row.split(','));

        renderDataGrid(dbName, tableName, rows);

    } catch (error) {
        console.error(error);
        mainView.innerHTML = `<div class="text-red-500 p-4">Error loading data: ${error.message}</div>`;
    }
}

function renderDataGrid(dbName, tableName, rows) {
    const mainView = document.getElementById('main-view');
    mainView.innerHTML = '';

    // 1. The Header (Title + Add Button)
    // We use 'flex' to push them apart, and better margins
    const header = document.createElement('div');
    header.className = 'flex items-center justify-between mb-6 px-1';
    header.innerHTML = `
        <div class="flex items-center">
            <button onclick="fetchTables('${dbName}')" class="text-gray-400 hover:text-white mr-4 transition flex items-center">
                <span class="text-xl mr-1">←</span> Back
            </button>
            <h2 class="text-3xl font-bold text-gray-100 flex items-center tracking-tight">
                <span class="text-blue-500 mr-3 text-2xl">📄</span> ${tableName}
            </h2>
            <span class="ml-4 text-xs font-mono text-blue-300 bg-blue-900 bg-opacity-30 px-2 py-1 rounded border border-blue-800">
                ${rows.length} records
            </span>
        </div>
        
        <button onclick="promptInsertRow('${dbName}', '${tableName}')" 
                class="bg-blue-600 hover:bg-blue-500 text-white px-5 py-2 rounded-lg text-sm font-medium shadow-lg transition-transform transform hover:scale-105 active:scale-95">
            + Add Row
        </button>
    `;
    mainView.appendChild(header);

    // 2. The Table Card (The Container)
    // This gives it the "Dashboard" look with a distinct background
    const tableContainer = document.createElement('div');
    tableContainer.className = 'bg-gray-800 rounded-xl shadow-2xl border border-gray-700 overflow-hidden';

    let html = `
        <div class="overflow-x-auto">
            <table class="w-full text-left text-sm text-gray-300">
                <thead class="bg-gray-900 text-gray-400 uppercase font-semibold tracking-wider">
                    <tr>
                        ${rows[0].map((_, i) => `<th class="px-6 py-4 border-b border-gray-700">Col ${i + 1}</th>`).join('')}
                        <th class="px-6 py-4 border-b border-gray-700 text-right">Actions</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-gray-700">
    `;

    // 3. The Rows
    rows.forEach(row => {
        // We assume Column 0 is the ID for the delete button
        const pk = row[0];
        
        html += `<tr class="hover:bg-gray-700 transition-colors duration-150 group">`;
        
        // Data Cells
        row.forEach(cell => {
            html += `<td class="px-6 py-4 whitespace-nowrap group-hover:text-white">${cell}</td>`;
        });

        // Delete Button (Only visible on hover)
        html += `
            <td class="px-6 py-4 text-right whitespace-nowrap">
                <button onclick="deleteRow('${dbName}', '${tableName}', '${pk}')" 
                        class="text-red-500 opacity-0 group-hover:opacity-100 hover:text-red-400 transition transform hover:scale-110 p-2" 
                        title="Delete Row">
                    🗑️
                </button>
            </td>
        </tr>`;
    });

    html += `</tbody></table></div>`;
    tableContainer.innerHTML = html;
    mainView.appendChild(tableContainer);
}

function renderEmptyTableState(dbName, tableName) {
    const mainView = document.getElementById('main-view');
    mainView.innerHTML = `
        <div class="flex flex-col h-full">
            <button onclick="fetchTables('${dbName}')" class="text-left text-gray-400 hover:text-white mb-4">← Back</button>
            <div class="flex-1 flex flex-col items-center justify-center text-gray-500">
                <p class="mb-4">Table is empty.</p>
                <button onclick="promptInsertRow('${dbName}', '${tableName}')" class="text-blue-400 hover:underline">
                    Add your first row
                </button>
            </div>
        </div>
    `;
}



async function handleInsert(dbName, tableName) {
    const values = prompt(`Enter values for ${tableName} (separated by comma):`, "3,NewUser");

    if (values === null) return;

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

        const result = await response.text();

        if (response.ok) {
            alert("Success!");
            fetchTableData(dbName, tableName); 
        } else {
            alert("Error: " + result);
        }

    } catch (error) {
        console.error("Insert failed:", error);
        alert("Failed to connect to server.");
    }
}



function deleteRow(dbName, tableName, pkValue) {

    Modal.open({
        title: 'Delete Row?',
        msg: `Are you sure you want to permanently delete ID "${pkValue}"? This action cannot be undone.`,
        showInput: false, 
        onConfirm: async () => {
            try {
                const response = await fetch('/data/delete', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        db_name: dbName,
                        table_name: tableName,
                        pk_value: pkValue
                    })
                });
                
                if (response.ok) {
                    fetchTableData(dbName, tableName);
                } else {
                    const err = await response.text();
                    alert("Error: " + err); 
                }
            } catch (error) {
                console.error(error);
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
                    alert("Error: " + err);
                }
            } catch (error) {
                console.error(error);
            }
        }
    });
}

fetchDatabases()