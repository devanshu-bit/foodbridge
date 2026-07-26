document.addEventListener('DOMContentLoaded', function() {
    authManager.updateNavbar();

    // Mobile menu toggle
    const hamburger = document.getElementById('hamburger');
    const navLinks = document.getElementById('navLinks');
    
    if (hamburger) {
        hamburger.addEventListener('click', () => navLinks.classList.toggle('active'));
    }

    // Logout
    const logoutBtn = document.getElementById('logoutBtn');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', (e) => {
            e.preventDefault();
            authManager.logout();
        });
    }

    // Close mobile menu on outside click
    document.addEventListener('click', (e) => {
        if (!hamburger?.contains(e.target) && !navLinks?.contains(e.target)) {
            navLinks?.classList.remove('active');
        }
    });

    // Smooth scroll
    document.querySelectorAll('a[href^="#"]').forEach(anchor => {
        anchor.addEventListener('click', function(e) {
            e.preventDefault();
            const target = document.querySelector(this.getAttribute('href'));
            if (target) {
                target.scrollIntoView({ behavior: 'smooth', block: 'start' });
            }
        });
    });
});

// Utility functions
function showAlert(message, type = 'info') {
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type}`;
    alertDiv.innerHTML = `<i class="fas fa-${type === 'success' ? 'check-circle' : type === 'error' ? 'exclamation-circle' : 'info-circle'}"></i> ${message}`;
    
    const container = document.querySelector('.dashboard-container') || document.querySelector('.auth-box');
    if (container) {
        container.insertBefore(alertDiv, container.firstChild);
        setTimeout(() => alertDiv.remove(), 5000);
    }
}

function formatDate(dateString) {
    return new Date(dateString).toLocaleString('en-IN', {
        year: 'numeric', month: 'short', day: 'numeric',
        hour: '2-digit', minute: '2-digit'
    });
}

function getStatusBadge(status) {
    const statusMap = {
        'available': 'status-available',
        'accepted': 'status-accepted',
        'assigned': 'status-assigned',
        'picked_up': 'status-picked_up',
        'delivered': 'status-delivered'
    };
    return `<span class="status-badge ${statusMap[status] || 'status-available'}">${status.replace('_', ' ').toUpperCase()}</span>`;
}

function showLoading(elementId) {
    document.getElementById(elementId).classList.add('active');
}

function hideLoading(elementId) {
    document.getElementById(elementId).classList.remove('active');
}