/**
 * cli_composer.js - Real-Time CLI Command Composer & Console Manager.
 *
 * Dynamically composes CLI command strings based on user actions.
 * Provides clipboard copying, direct execution, and terminal log stream viewing.
 */

class CliComposer {
    constructor() {
        this.panel = document.getElementById('cli-panel');
        this.cmdPreview = document.getElementById('cli-cmd-preview');
        this.outputBox = document.getElementById('cli-output-box');
        this.toggleBtn = document.getElementById('cli-toggle-btn');
        this.copyBtn = document.getElementById('cli-copy-btn');
        this.runBtn = document.getElementById('cli-run-btn');
        this.clearBtn = document.getElementById('cli-clear-btn');

        this.currentCommand = 'f108-pro --help';
        this.isExpanded = true;
    }

    init() {
        if (!this.panel) return;

        // Toggle expand / collapse
        if (this.toggleBtn) {
            this.toggleBtn.addEventListener('click', () => this.toggle());
        }

        // Copy command button
        if (this.copyBtn) {
            this.copyBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.copyCommand();
            });
        }

        // Run command button
        if (this.runBtn) {
            this.runBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.runCurrentCommand();
            });
        }

        // Clear terminal output
        if (this.clearBtn) {
            this.clearBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.clearOutput();
            });
        }

        // Global hotkey: Ctrl + ` (backtick) toggles console
        window.addEventListener('keydown', (e) => {
            if (e.ctrlKey && e.key === '`') {
                e.preventDefault();
                this.toggle();
            }
        });

        // Listen for log streams from hostBridge
        window.hostBridge.on('cli_log', (data) => {
            this.appendLog(data.text, data.stream);
        });

        this.setCommand('f108-pro clock');
    }

    /**
     * Updates the active command being composed.
     * @param {string} cmdString
     */
    setCommand(cmdString) {
        this.currentCommand = cmdString.trim();
        if (this.cmdPreview) {
            this.cmdPreview.textContent = this.currentCommand;
        }
    }

    /**
     * Copies current command to system clipboard.
     */
    async copyCommand() {
        try {
            await navigator.clipboard.writeText(this.currentCommand);
            const originalText = this.copyBtn.innerHTML;
            this.copyBtn.textContent = 'Copied!';
            this.copyBtn.style.background = 'var(--success)';
            setTimeout(() => {
                this.copyBtn.innerHTML = originalText;
                this.copyBtn.style.background = '';
            }, 1200);
        } catch (err) {
            console.error('Clipboard copy failed:', err);
        }
    }

    /**
     * Runs current command asynchronously via the IPC bridge.
     */
    async runCurrentCommand() {
        if (!this.currentCommand) return;

        if (!this.isExpanded) {
            this.expand();
        }

        try {
            this.runBtn.disabled = true;
            this.runBtn.style.opacity = '0.6';
            this.appendLog('[INFO] Comando de referencia: usa el botón Apply de la pestaña.' + String.fromCharCode(10), 'info');
        } catch (err) {
            this.appendLog(`[ERROR] ${err.message}\n`, 'stderr');
        } finally {
            this.runBtn.disabled = false;
            this.runBtn.style.opacity = '1';
        }
    }

    /**
     * Appends a log line to the terminal output area.
     */
    appendLog(text, stream = 'stdout') {
        if (!this.outputBox) return;
        const span = document.createElement('span');
        span.className = `cli-line-${stream}`;
        span.textContent = text;
        this.outputBox.appendChild(span);
        this.outputBox.scrollTop = this.outputBox.scrollHeight;
    }

    /**
     * Clears the terminal output stream.
     */
    clearOutput() {
        if (this.outputBox) {
            this.outputBox.innerHTML = '';
        }
    }

    /**
     * Toggles the collapsible bottom console.
     */
    toggle() {
        if (this.isExpanded) {
            this.collapse();
        } else {
            this.expand();
        }
    }

    expand() {
        this.isExpanded = true;
        this.panel.classList.remove('collapsed');
        this.panel.classList.add('expanded');
    }

    collapse() {
        this.isExpanded = false;
        this.panel.classList.remove('expanded');
        this.panel.classList.add('collapsed');
    }
}

window.CliComposer = CliComposer;
