function filterTable(tableId, query) {
    const rows = document.querySelectorAll('#' + tableId + ' tbody tr');
    const q = query.toLowerCase().trim();
    rows.forEach(row => { row.style.display = row.textContent.toLowerCase().includes(q) ? '' : 'none'; });
}
