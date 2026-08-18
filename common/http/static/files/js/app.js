(function() {
    const body = document.body;
    const html = document.documentElement;
    const KEY_W = 'sidebar-w';
    const KEY_COLLAPSED = 'sidebar-collapsed';
    const MIN_W = 200, MAX_W = 380, DEFAULT_W = 248;

    if (html.getAttribute('data-pre-collapsed') === 'true') {
        body.setAttribute('data-sidebar-collapsed', 'true');
        html.removeAttribute('data-pre-collapsed');
    }

    function toggleCollapse() {
        const collapsed = body.getAttribute('data-sidebar-collapsed') === 'true';
        if (collapsed) {
            body.removeAttribute('data-sidebar-collapsed');
            localStorage.removeItem(KEY_COLLAPSED);
        } else {
            body.setAttribute('data-sidebar-collapsed', 'true');
            localStorage.setItem(KEY_COLLAPSED, '1');
        }
    }
    document.querySelectorAll('[data-action="sidebar-toggle"]').forEach(el =>
        el.addEventListener('click', toggleCollapse)
    );

    document.addEventListener('keydown', e => {
        if ((e.metaKey || e.ctrlKey) && e.key === '\\') {
            e.preventDefault();
            toggleCollapse();
        }
        if (e.key === 'Escape') body.removeAttribute('data-sidebar-open');
    });

    document.querySelectorAll('[data-action="sidebar-open"]').forEach(el =>
        el.addEventListener('click', () => body.setAttribute('data-sidebar-open', 'true'))
    );
    document.querySelectorAll('[data-action="sidebar-close"]').forEach(el =>
        el.addEventListener('click', () => body.removeAttribute('data-sidebar-open'))
    );

    const handle = document.querySelector('[data-action="sidebar-resize"]');
    if (handle) {
        let startX = 0, startW = DEFAULT_W;

        const onMove = e => {
            const x = e.touches ? e.touches[0].clientX : e.clientX;
            const w = Math.min(MAX_W, Math.max(MIN_W, startW + (x - startX)));
            html.style.setProperty('--sidebar-w', w + 'px');
        };

        const onUp = () => {
            body.removeAttribute('data-sidebar-resizing');
            const w = html.style.getPropertyValue('--sidebar-w');
            if (w) localStorage.setItem(KEY_W, w.trim());
            document.removeEventListener('mousemove', onMove);
            document.removeEventListener('mouseup', onUp);
            document.removeEventListener('touchmove', onMove);
            document.removeEventListener('touchend', onUp);
        };

        const onDown = e => {
            if (body.getAttribute('data-sidebar-collapsed') === 'true') return;
            startX = e.touches ? e.touches[0].clientX : e.clientX;
            const cur = getComputedStyle(html).getPropertyValue('--sidebar-w').trim();
            startW = parseInt(cur) || DEFAULT_W;
            body.setAttribute('data-sidebar-resizing', 'true');
            document.addEventListener('mousemove', onMove);
            document.addEventListener('mouseup', onUp);
            document.addEventListener('touchmove', onMove, { passive: false });
            document.addEventListener('touchend', onUp);
            e.preventDefault();
        };

        handle.addEventListener('mousedown', onDown);
        handle.addEventListener('touchstart', onDown, { passive: false });

        handle.addEventListener('dblclick', () => {
            html.style.removeProperty('--sidebar-w');
            localStorage.removeItem(KEY_W);
        });
    }

    document.querySelectorAll('.sidebar-nav a').forEach(a =>
        a.addEventListener('click', () => body.removeAttribute('data-sidebar-open'))
    );

    const userMenus = document.querySelectorAll('details.user-menu');
    document.addEventListener('click', e => {
        userMenus.forEach(d => {
            if (d.hasAttribute('open') && !d.contains(e.target)) d.removeAttribute('open');
        });
    });
    document.addEventListener('keydown', e => {
        if (e.key === 'Escape') userMenus.forEach(d => d.removeAttribute('open'));
    });
    document.querySelectorAll('.user-menu-item').forEach(item =>
        item.addEventListener('click', () => {
            if (item.getAttribute('data-action') !== 'theme-toggle') {
                item.closest('details.user-menu')?.removeAttribute('open');
            }
        })
    );

    function setThemeLabel() {
        const isDark = html.getAttribute('data-theme') === 'dark';
        document.querySelectorAll('[data-theme-label]').forEach(el => {
            el.textContent = isDark ? 'Oscuro' : 'Claro';
        });
    }
    function toggleTheme() {
        const isDark = html.getAttribute('data-theme') === 'dark';
        if (isDark) {
            html.removeAttribute('data-theme');
            localStorage.setItem('theme', 'light');
        } else {
            html.setAttribute('data-theme', 'dark');
            localStorage.setItem('theme', 'dark');
        }
        setThemeLabel();
    }
    document.querySelectorAll('[data-action="theme-toggle"]').forEach(el =>
        el.addEventListener('click', e => {
            e.preventDefault();
            toggleTheme();
        })
    );
    setThemeLabel();

    if (window.matchMedia) {
        window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', e => {
            if (!localStorage.getItem('theme')) {
                if (e.matches) html.setAttribute('data-theme', 'dark');
                else html.removeAttribute('data-theme');
                setThemeLabel();
            }
        });
    }

    document.body.addEventListener('htmx:beforeSwap', evt => {
        const status = evt.detail.xhr.status;
        if (status === 400 || status === 422) {
            evt.detail.shouldSwap = true;
            evt.detail.isError = false;
        }
    });

    const modal = document.getElementById('modal');
    function openModal() {
        if (!modal) return;
        modal.setAttribute('data-state', 'open');
        modal.setAttribute('aria-hidden', 'false');
        body.setAttribute('data-modal-open', 'true');
    }
    function closeModal() {
        if (!modal) return;
        modal.setAttribute('data-state', 'closed');
        modal.setAttribute('aria-hidden', 'true');
        modal.innerHTML = '';
        body.removeAttribute('data-modal-open');
    }
    if (modal) {
        modal.addEventListener('htmx:afterSwap', () => {
            if (modal.innerHTML.trim() !== '') openModal();
        });
        modal.addEventListener('click', e => {
            if (e.target === modal || e.target.closest('[data-action="close-modal"]')) closeModal();
        });
        document.body.addEventListener('closeModal', closeModal);
        document.addEventListener('keydown', e => {
            if (e.key === 'Escape' && modal.getAttribute('data-state') === 'open') closeModal();
        });
    }
})();
