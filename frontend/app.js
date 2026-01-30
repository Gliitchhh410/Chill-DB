const dbListContainer = document.getElementById("db-list")
let currentDB = null;


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
            // Tailwind: dark bg, border, hover effect
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
            card.onclick = () => alert(`Viewing data for ${tableName} (Coming Soon)`);

            grid.appendChild(card);
        });
        

        mainView.appendChild(grid);
    } catch (error) {
        console.error(error);
        mainView.innerHTML = `<div class="text-red-500">Error loading tables: ${error.message}</div>`;
    }
}


fetchDatabases()