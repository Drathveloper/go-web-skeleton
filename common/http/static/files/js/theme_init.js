(function() {
    try {
        var w = localStorage.getItem('sidebar-w');
        if (w) document.documentElement.style.setProperty('--sidebar-w', w);
        if (localStorage.getItem('sidebar-collapsed') === '1') {
            document.documentElement.setAttribute('data-pre-collapsed', 'true');
        }
        var saved = localStorage.getItem('theme');
        var dark = saved
            ? saved === 'dark'
            : window.matchMedia && window.matchMedia('(prefers-color-scheme: dark)').matches;
        if (dark) document.documentElement.setAttribute('data-theme', 'dark');
    } catch(e) {}
})();
