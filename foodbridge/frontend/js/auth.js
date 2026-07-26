class AuthManager {
    constructor() {
        this.API_URL = 'http://localhost:8080/api';
        this.token = localStorage.getItem('foodbridge_token');
        this.user = JSON.parse(localStorage.getItem('foodbridge_user') || 'null');
    }

    async register(userData) {
        const response = await fetch(`${this.API_URL}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(userData)
        });
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || 'Registration failed');
        this.setSession(data.token, data.user);
        return data;
    }

    async login(email, password) {
        const response = await fetch(`${this.API_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ email, password })
        });
        const data = await response.json();
        if (!response.ok) throw new Error(data.error || 'Login failed');
        this.setSession(data.token, data.user);
        return data;
    }

    setSession(token, user) {
        this.token = token;
        this.user = user;
        localStorage.setItem('foodbridge_token', token);
        localStorage.setItem('foodbridge_user', JSON.stringify(user));
    }

    logout() {
        this.token = null;
        this.user = null;
        localStorage.removeItem('foodbridge_token');
        localStorage.removeItem('foodbridge_user');
        window.location.href = '/index.html';
    }

    isAuthenticated() {
        return !!this.token;
    }

    getToken() {
        return this.token;
    }

    getUser() {
        return this.user;
    }

    getDashboardUrl() {
        if (!this.user) return '/';
        switch (this.user.role) {
            case 'restaurant': return '/pages/restaurant-dashboard.html';
            case 'ngo':        return '/pages/ngo-dashboard.html';
            case 'volunteer':  return '/pages/volunteer-dashboard.html';
            case 'admin':      return '/pages/admin-dashboard.html';
            default:           return '/';
        }
    }

    updateNavbar() {
        const loginLink = document.getElementById('loginLink');
        const registerLink = document.getElementById('registerLink');
        const dashboardLink = document.getElementById('dashboardLink');
        const logoutBtn = document.getElementById('logoutBtn');
        if (!loginLink) return;
        if (this.isAuthenticated()) {
            if (loginLink) loginLink.style.display = 'none';
            if (registerLink) registerLink.style.display = 'none';
            if (dashboardLink) {
                dashboardLink.style.display = 'inline-block';
                dashboardLink.href = this.getDashboardUrl();
            }
            if (logoutBtn) logoutBtn.style.display = 'inline-block';
        } else {
            if (loginLink) loginLink.style.display = 'inline-block';
            if (registerLink) registerLink.style.display = 'inline-block';
            if (dashboardLink) dashboardLink.style.display = 'none';
            if (logoutBtn) logoutBtn.style.display = 'none';
        }
    }
}

const authManager = new AuthManager();

async function apiCall(endpoint, method = 'GET', body = null) {
    const headers = { 'Content-Type': 'application/json' };
    if (authManager.getToken()) headers['Authorization'] = `Bearer ${authManager.getToken()}`;
    const config = { method, headers };
    if (body) config.body = JSON.stringify(body);
    const response = await fetch(`${authManager.API_URL}/${endpoint}`, config);
    const data = await response.json();
    if (!response.ok) throw new Error(data.error || 'API call failed');
    return data;
}

function showAlert(message, type = 'error') {
    const existingAlert = document.querySelector('.alert');
    if (existingAlert) existingAlert.remove();
    const alertDiv = document.createElement('div');
    alertDiv.className = `alert alert-${type}`;
    alertDiv.textContent = message;
    const form = document.querySelector('form');
    if (form) {
        form.parentNode.insertBefore(alertDiv, form);
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
    const classes = {
        'available': 'status-available',
        'accepted': 'status-accepted',
        'assigned': 'status-assigned',
        'picked_up': 'status-picked_up',
        'delivered': 'status-delivered'
    };
    return `<span class="status-badge ${classes[status]}">${status.replace('_', ' ')}</span>`;
}

function showLoading() {
    const loading = document.getElementById('loading');
    if (loading) loading.classList.add('active');
}

function hideLoading() {
    const loading = document.getElementById('loading');
    if (loading) loading.classList.remove('active');
}