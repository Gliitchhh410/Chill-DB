const dbListContainer = document.getElementById("db-list")



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
                alert(`You clicked on ${db}`)
            }

            dbListContainer.appendChild(item)
        });
    } catch (e){
        console.error(`Error fetching the databases: ${e}`);
        dbListContainer.innerHTML = '<div class="text-red-500 p-2 text-xs">Failed to load</div>'
    }
}


fetchDatabases()