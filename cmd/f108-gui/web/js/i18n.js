/**
 * i18n.js - Internationalization Engine for AULA F108 Pro GUI.
 *
 * Supports dynamic locale switching (en, pt_BR, and future community translations).
 * Replaces text content of elements with [data-i18n] and placeholders with [data-i18n-placeholder].
 */

class I18nEngine {
    constructor() {
        this.currentLocale = localStorage.getItem('app_locale_v3') || 'es';
        this.translations = {};
        this.loadedLocales = new Map();
    }

    /**
     * Initializes the i18n engine with the saved or default locale.
     */
    async init() {
        await this.setLocale(this.currentLocale);
    }

    /**
     * Loads a locale JSON file and updates the DOM.
     * @param {string} locale - Locale code (e.g. 'en', 'pt_BR')
     */
    async setLocale(locale) {
        if (!this.loadedLocales.has(locale)) {
            try {
                const res = await fetch(`locales/${locale}.json`);
                // Note: file:// protocol returns status 0, ok false
                if (res.status !== 0 && !res.ok) {
                    throw new Error(`HTTP error ${res.status}`);
                }
                const data = await res.json();
                this.loadedLocales.set(locale, data);
            } catch (err) {
                console.error(`Failed to load locale '${locale}', falling back to 'en':`, err);
                if (locale !== 'en') {
                    return this.setLocale('en');
                }
            }
        }

        this.currentLocale = locale;
        this.translations = this.loadedLocales.get(locale) || {};
        localStorage.setItem('app_locale_v3', locale);
        this.applyTranslations();

        // Dispatch custom event for modules that need to re-render dynamic strings
        window.dispatchEvent(new CustomEvent('localeChanged', { detail: { locale } }));
    }

    /**
     * Translates a dot-notated key (e.g. 'nav.home', 'lighting.modes.Breath').
     * @param {string} key
     * @param {string} defaultVal
     * @returns {string}
     */
    t(key, defaultVal = '') {
        const parts = key.split('.');
        let curr = this.translations;
        for (const p of parts) {
            if (curr && typeof curr === 'object' && p in curr) {
                curr = curr[p];
            } else {
                return defaultVal || key;
            }
        }
        return typeof curr === 'string' ? curr : (defaultVal || key);
    }

    /**
     * Scans document for data-i18n and data-i18n-placeholder attributes and updates them.
     */
    applyTranslations(root = document) {
        // Translate text content
        const elements = root.querySelectorAll('[data-i18n]');
        elements.forEach(el => {
            const key = el.getAttribute('data-i18n');
            const translation = this.t(key);
            if (translation) {
                el.textContent = translation;
            }
        });

        // Translate placeholders (inputs, textareas)
        const placeholders = root.querySelectorAll('[data-i18n-placeholder]');
        placeholders.forEach(el => {
            const key = el.getAttribute('data-i18n-placeholder');
            const translation = this.t(key);
            if (translation) {
                el.setAttribute('placeholder', translation);
            }
        });

        // Translate title tooltips
        const titles = root.querySelectorAll('[data-i18n-title]');
        titles.forEach(el => {
            const key = el.getAttribute('data-i18n-title');
            const translation = this.t(key);
            if (translation) {
                el.setAttribute('title', translation);
            }
        });
    }
}

// Global instance
window.i18n = new I18nEngine();
