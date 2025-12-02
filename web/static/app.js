const API_BASE = 'http://localhost:8080/api';
let currentFilter = 'all';
let allJobs = [];

// Load data on page load
document.addEventListener('DOMContentLoaded', () => {
    loadConfig();
    loadJobs();
    loadStats();

    // Auto-refresh every 5 seconds
    setInterval(() => {
        loadJobs();
        loadStats();
    }, 5000);
});

async function loadConfig() {
    try {
        const response = await fetch(`${API_BASE}/config`);
        const data = await response.json();
        document.getElementById('maxConcurrent').value = data.max_concurrent_jobs;
    } catch (error) {
        console.error('Failed to load config:', error);
    }
}

async function updateConfig() {
    const maxConcurrent = parseInt(document.getElementById('maxConcurrent').value);

    if (maxConcurrent < 1) {
        alert('Max concurrent jobs must be at least 1');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/config`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ max_concurrent_jobs: maxConcurrent })
        });

        if (response.ok) {
            showNotification('Config updated successfully!', 'success');
        } else {
            showNotification('Failed to update config', 'error');
        }
    } catch (error) {
        console.error('Failed to update config:', error);
        showNotification('Failed to update config', 'error');
    }
}

async function loadStats() {
    try {
        const response = await fetch(`${API_BASE}/stats`);
        const data = await response.json();

        document.getElementById('queuedCount').textContent = data.queued_count;
        document.getElementById('runningCount').textContent = data.running_count;
        document.getElementById('completedCount').textContent = data.completed_count;
        document.getElementById('failedCount').textContent = data.failed_count;
    } catch (error) {
        console.error('Failed to load stats:', error);
    }
}

async function loadJobs() {
    try {
        const response = await fetch(`${API_BASE}/jobs`);
        allJobs = await response.json();
        renderJobs();
    } catch (error) {
        console.error('Failed to load jobs:', error);
        document.getElementById('jobsBody').innerHTML =
            '<tr><td colspan="6" class="loading">Failed to load jobs</td></tr>';
    }
}

function renderJobs() {
    const tbody = document.getElementById('jobsBody');

    let filteredJobs = allJobs;
    if (currentFilter !== 'all') {
        filteredJobs = allJobs.filter(job => job.status === currentFilter);
    }

    if (filteredJobs.length === 0) {
        tbody.innerHTML = '<tr><td colspan="6" class="loading">No jobs found</td></tr>';
        return;
    }

    tbody.innerHTML = filteredJobs.map(job => `
        <tr>
            <td>${job.id}</td>
            <td><strong>${job.name}</strong></td>
            <td><span class="command" title="${escapeHtml(job.command)}">${escapeHtml(job.command)}</span></td>
            <td><span class="status ${job.status}">${job.status}</span></td>
            <td class="timestamp">${formatTimestamp(job.last_run)}</td>
            <td class="timestamp">${formatTimestamp(job.created_at)}</td>
        </tr>
    `).join('');
}

function filterJobs(status) {
    currentFilter = status;

    // Update active tab
    document.querySelectorAll('.tab').forEach(tab => {
        tab.classList.remove('active');
    });
    event.target.classList.add('active');

    renderJobs();
}

async function createJob(event) {
    event.preventDefault();

    const name = document.getElementById('jobName').value.trim();
    const command = document.getElementById('jobCommand').value.trim();

    if (!name || !command) {
        alert('Please fill in all fields');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/jobs/create`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, command })
        });

        if (response.ok) {
            showNotification('Job created successfully!', 'success');
            document.getElementById('jobForm').reset();
            loadJobs();
            loadStats();
        } else {
            const error = await response.text();
            showNotification(`Failed to create job: ${error}`, 'error');
        }
    } catch (error) {
        console.error('Failed to create job:', error);
        showNotification('Failed to create job', 'error');
    }
}

function refreshData() {
    loadConfig();
    loadJobs();
    loadStats();
    showNotification('Data refreshed!', 'success');
}

function formatTimestamp(timestamp) {
    if (!timestamp) return 'Never';
    const date = new Date(timestamp);
    return date.toLocaleString();
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function showNotification(message, type = 'success') {
    // Simple notification - you can enhance this later
    const notification = document.createElement('div');
    notification.textContent = message;
    notification.style.cssText = `
        position: fixed;
        top: 20px;
        right: 20px;
        padding: 15px 25px;
        background: ${type === 'success' ? '#10b981' : '#ef4444'};
        color: white;
        border-radius: 8px;
        box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        z-index: 1000;
        animation: slideIn 0.3s ease;
    `;

    document.body.appendChild(notification);

    setTimeout(() => {
        notification.style.animation = 'slideOut 0.3s ease';
        setTimeout(() => notification.remove(), 300);
    }, 3000);
}

// Add CSS animations
const style = document.createElement('style');
style.textContent = `
    @keyframes slideIn {
        from {
            transform: translateX(400px);
            opacity: 0;
        }
        to {
            transform: translateX(0);
            opacity: 1;
        }
    }

    @keyframes slideOut {
        from {
            transform: translateX(0);
            opacity: 1;
        }
        to {
            transform: translateX(400px);
            opacity: 0;
        }
    }
`;
document.head.appendChild(style);
