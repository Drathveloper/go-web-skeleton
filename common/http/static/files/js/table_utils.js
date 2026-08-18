// Helpers shared by every listing rendered through the table components.
//
// Filtering and sorting are client side, over the rows already in the DOM: the
// listings are small and the alternative is a round trip per keystroke. The
// day a listing outgrows this, the fix is server side pagination in the table
// components, not a change in each generated page.

// filterTable hides the rows of tableId that do not contain query. The empty
// state row is left to syncEmptyRow, which decides on its own whether anything
// is visible.
function filterTable(tableId, query) {
    const table = document.getElementById(tableId);
    if (!table) return;
    const q = query.toLowerCase().trim();
    table.querySelectorAll('tbody tr').forEach(row => {
        if (row.hasAttribute('data-empty-row')) return;
        row.hidden = q !== '' && !row.textContent.toLowerCase().includes(q);
    });
    syncEmptyRow(table);
}

// syncEmptyRow shows the "no results" row only while no data row is visible,
// which also covers the row an out-of-band swap has just prepended.
function syncEmptyRow(table) {
    const empty = table.querySelector('tbody tr[data-empty-row]');
    if (!empty) return;
    const visible = Array.from(table.querySelectorAll('tbody tr'))
        .some(row => !row.hasAttribute('data-empty-row') && !row.hidden);
    empty.hidden = visible;
}

// sortTable reorders the rows of the table by the text of column index, keeping
// the empty state row at the bottom. Numeric columns are compared as numbers so
// that 10 does not sort before 9.
function sortTable(table, index, ascending) {
    const body = table.querySelector('tbody');
    if (!body) return;
    const rows = Array.from(body.querySelectorAll('tr')).filter(row => !row.hasAttribute('data-empty-row'));
    const value = row => (row.children[index] ? row.children[index].textContent.trim() : '');
    const numeric = rows.every(row => value(row) === '' || !isNaN(parseFloat(value(row))));
    rows.sort((left, right) => {
        const a = value(left), b = value(right);
        const result = numeric ? parseFloat(a || '0') - parseFloat(b || '0') : a.localeCompare(b);
        return ascending ? result : -result;
    });
    rows.forEach(row => body.appendChild(row));
    const empty = body.querySelector('tr[data-empty-row]');
    if (empty) body.appendChild(empty);
}

// initDataTable wires the sortable headers of one table. Called once per table
// by the table/script component, which is the only per table javascript in the
// application.
function initDataTable(tableId) {
    const table = document.getElementById(tableId);
    if (!table || table.dataset.tableReady === 'true') return;
    table.dataset.tableReady = 'true';

    table.querySelectorAll('thead th').forEach((th, index) => {
        const control = th.querySelector('[data-sort-key]');
        if (!control) return;
        control.addEventListener('click', () => {
            const ascending = th.getAttribute('aria-sort') !== 'ascending';
            table.querySelectorAll('thead th[aria-sort]').forEach(other => other.removeAttribute('aria-sort'));
            th.setAttribute('aria-sort', ascending ? 'ascending' : 'descending');
            sortTable(table, index, ascending);
        });
    });

    syncEmptyRow(table);
}

// An out-of-band swap adds or removes a row without going through any of the
// functions above, so the empty state is re-evaluated after every swap.
document.addEventListener('htmx:afterSwap', () => {
    document.querySelectorAll('table tbody tr[data-empty-row]')
        .forEach(row => syncEmptyRow(row.closest('table')));
});
