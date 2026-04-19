// Modal management
function showModal() {
    document.getElementById('modal-backdrop').classList.add('show');
}

function hideModal() {
    document.getElementById('modal-backdrop').classList.remove('show');
    document.getElementById('modal-content').innerHTML = '';
}

function showConfirmModal(name, deleteUrl, target) {
    document.getElementById('confirm-item-name').textContent = name;
    const btn = document.getElementById('confirm-delete-btn');
    btn.setAttribute('hx-delete', deleteUrl);
    btn.setAttribute('hx-target', target || '#services-table');
    htmx.process(btn);
    document.getElementById('confirm-modal').classList.add('show');
}

function hideConfirmModal() {
    document.getElementById('confirm-modal').classList.remove('show');
}

// Close modal on ESC
document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') {
        hideModal();
        hideConfirmModal();
    }
});

// HTMX: open modal when modal-content is loaded
document.body.addEventListener('htmx:afterSwap', function(e) {
    if (e.detail.target.id === 'modal-content') {
        showModal();
    }
});

// HTMX: close modal after successful form submit
document.body.addEventListener('serviceUpdated', hideModal);
document.body.addEventListener('categoryUpdated', hideModal);

// Confirm modal hide after delete
document.body.addEventListener('htmx:afterRequest', function(e) {
    if (e.detail.elt && e.detail.elt.id === 'confirm-delete-btn') {
        hideConfirmModal();
    }
});

// Color input sync: keep hex text field in sync with color picker
document.addEventListener('input', function(e) {
    if (e.target.type === 'color' && e.target.dataset.syncTo) {
        const target = document.getElementById(e.target.dataset.syncTo);
        if (target) target.value = e.target.value;
    }
    if (e.target.id === 'color-text' && e.target.dataset.syncTo) {
        const target = document.getElementById(e.target.dataset.syncTo);
        if (target && /^#[0-9a-fA-F]{6}$/.test(e.target.value)) {
            target.value = e.target.value;
        }
    }
});
