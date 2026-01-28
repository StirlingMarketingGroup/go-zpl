// ZPL Label Renderer - Client-side application

let editor;
let renderTimeout;
let wasmReady = false;

// Storage keys
const STORAGE_KEY = 'zpl-renderer';

// Default 4x6 label at 203 DPI
const DEFAULT_DPI = 203;
const DEFAULT_WIDTH_INCHES = 4;
const DEFAULT_HEIGHT_INCHES = 6;

const defaultZPL = `^XA
^FO50,50
^A0N,30,30
^FDHello, World!^FS

^FO50,100
^BQN,2,5
^FDMM,Ahttps://github.com/StirlingMarketingGroup/go-zpl^FS

^FO250,50
^A0N,25,25
^FDgo-zpl WASM Demo^FS

^XZ`;

// Load saved state from localStorage
function loadState() {
    try {
        const saved = localStorage.getItem(STORAGE_KEY);
        if (saved) {
            return JSON.parse(saved);
        }
    } catch (e) {
        console.warn('Failed to load saved state:', e);
    }
    return null;
}

// Save state to localStorage
function saveState() {
    try {
        const state = {
            dpi: document.getElementById('dpi').value,
            width: document.getElementById('width').value,
            height: document.getElementById('height').value,
            unit: document.getElementById('unit').value,
            ignoreLabelHome: document.getElementById('ignore-label-home').checked,
            zpl: editor ? editor.getValue() : defaultZPL,
        };
        localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } catch (e) {
        console.warn('Failed to save state:', e);
    }
}

// Convert between dots and display units
function dotsToDisplay(dots, dpi, unit) {
    if (unit === 'in') {
        return (dots / dpi).toFixed(2);
    } else if (unit === 'mm') {
        return ((dots / dpi) * 25.4).toFixed(1);
    }
    return dots;
}

function displayToDots(value, dpi, unit) {
    if (unit === 'in') {
        return Math.round(parseFloat(value) * dpi);
    } else if (unit === 'mm') {
        return Math.round((parseFloat(value) / 25.4) * dpi);
    }
    return parseInt(value, 10) || 0;
}

// Update dimension displays when unit changes
function updateDimensionDisplays() {
    const dpi = parseInt(document.getElementById('dpi').value, 10);
    const unit = document.getElementById('unit').value;
    const widthEl = document.getElementById('width');
    const heightEl = document.getElementById('height');

    // Get current values in dots
    const widthDots = parseInt(widthEl.dataset.dots, 10) || displayToDots(widthEl.value, dpi, widthEl.dataset.unit || 'dots');
    const heightDots = parseInt(heightEl.dataset.dots, 10) || displayToDots(heightEl.value, dpi, heightEl.dataset.unit || 'dots');

    // Store dots value and update display
    widthEl.dataset.dots = widthDots;
    heightEl.dataset.dots = heightDots;
    widthEl.dataset.unit = unit;
    heightEl.dataset.unit = unit;

    if (unit === 'dots') {
        widthEl.value = widthDots;
        heightEl.value = heightDots;
        widthEl.step = '1';
        heightEl.step = '1';
    } else if (unit === 'in') {
        widthEl.value = dotsToDisplay(widthDots, dpi, unit);
        heightEl.value = dotsToDisplay(heightDots, dpi, unit);
        widthEl.step = '0.1';
        heightEl.step = '0.1';
    } else if (unit === 'mm') {
        widthEl.value = dotsToDisplay(widthDots, dpi, unit);
        heightEl.value = dotsToDisplay(heightDots, dpi, unit);
        widthEl.step = '1';
        heightEl.step = '1';
    }

    // Update labels
    document.getElementById('width-label').textContent = `Width (${unit}):`;
    document.getElementById('height-label').textContent = `Height (${unit}):`;
}

// Get dimensions in dots for rendering
function getDimensionsInDots() {
    const dpi = parseInt(document.getElementById('dpi').value, 10);
    const unit = document.getElementById('unit').value;
    const width = document.getElementById('width').value;
    const height = document.getElementById('height').value;

    return {
        width: displayToDots(width, dpi, unit),
        height: displayToDots(height, dpi, unit),
    };
}

// Initialize Monaco Editor
require.config({ paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs' } });

require(['vs/editor/editor.main'], function () {
    // Register ZPL language
    monaco.languages.register({ id: 'zpl' });

    // ZPL syntax highlighting
    monaco.languages.setMonarchTokensProvider('zpl', {
        tokenizer: {
            root: [
                [/\^XA/, 'keyword'],
                [/\^XZ/, 'keyword'],
                [/\^FO\d*,\d*/, 'variable'],
                [/\^FD[^^]*/, 'string'],
                [/\^FS/, 'delimiter'],
                [/\^A[0-9A-Z][NRIB]?,\d*,\d*/, 'type'],
                [/\^B[A-Z0-9]+/, 'function'],
                [/\^CF[A-Z0-9],\d*,?\d*/, 'type'],
                [/\^GB\d*,\d*,\d*,?[BWN]?,?\d*/, 'variable'],
                [/\^GF[ABC],\d*,\d*,\d*,/, 'function'],
                [/\^BY\d*,?\d*\.?\d*,?\d*/, 'variable'],
                [/\^PQ\d*,?\d*,?\d*,?[YN]?/, 'keyword'],
                [/\^LL\d*/, 'keyword'],
                [/\^LH\d*,\d*/, 'keyword'],
                [/\^CI\d*/, 'keyword'],
                [/\^[A-Z][A-Z0-9]*/, 'tag'],
                [/~[A-Z][A-Z0-9]*/, 'tag'],
            ],
        },
    });

    // Load saved state
    const saved = loadState();
    const initialZPL = saved?.zpl || defaultZPL;

    // Create editor
    editor = monaco.editor.create(document.getElementById('editor'), {
        value: initialZPL,
        language: 'zpl',
        theme: 'vs-dark',
        minimap: { enabled: false },
        automaticLayout: true,
        fontSize: 14,
        lineNumbers: 'on',
        scrollBeyondLastLine: false,
        wordWrap: 'on',
    });

    // Render on content change (debounced)
    editor.onDidChangeModelContent(() => {
        clearTimeout(renderTimeout);
        renderTimeout = setTimeout(() => {
            render();
            saveState();
        }, 300);
    });

    // Initial render when WASM is ready
    if (wasmReady) {
        render();
    }
});

// Initialize WASM
async function initWasm() {
    const go = new Go();
    try {
        const result = await WebAssembly.instantiateStreaming(
            fetch('lib.wasm'),
            go.importObject
        );
        go.run(result.instance);
        wasmReady = true;
        document.getElementById('loading').style.display = 'none';
        if (editor) {
            render();
        }
    } catch (err) {
        document.getElementById('loading').textContent = 'Failed to load WASM: ' + err.message;
        console.error('WASM load error:', err);
    }
}

// Render ZPL to PNG
function render() {
    if (!wasmReady || !editor) return;

    const zpl = editor.getValue();
    const dpi = parseInt(document.getElementById('dpi').value, 10);
    const dims = getDimensionsInDots();

    const errorEl = document.getElementById('error');
    const previewEl = document.getElementById('preview');
    const renderTimeEl = document.getElementById('render-time');

    const startTime = performance.now();

    try {
        const ignoreLabelHome = document.getElementById('ignore-label-home').checked;
        const result = window.renderZPL(zpl, dpi, dims.width, dims.height, ignoreLabelHome);
        const elapsed = (performance.now() - startTime).toFixed(1);

        if (result.error) {
            errorEl.textContent = result.error;
            errorEl.style.display = 'block';
            previewEl.style.display = 'none';
            renderTimeEl.textContent = '';
        } else {
            errorEl.style.display = 'none';
            previewEl.src = 'data:image/png;base64,' + result.image;
            previewEl.style.display = 'block';
            renderTimeEl.textContent = `${result.width}x${result.height} rendered in ${elapsed}ms`;
        }
    } catch (err) {
        errorEl.textContent = 'Render error: ' + err.message;
        errorEl.style.display = 'block';
        previewEl.style.display = 'none';
        renderTimeEl.textContent = '';
    }
}

// Initialize controls from saved state or defaults
function initControls() {
    const saved = loadState();
    const dpiEl = document.getElementById('dpi');
    const widthEl = document.getElementById('width');
    const heightEl = document.getElementById('height');
    const unitEl = document.getElementById('unit');
    const ignoreLabelHomeEl = document.getElementById('ignore-label-home');

    if (saved) {
        dpiEl.value = saved.dpi || DEFAULT_DPI;
        unitEl.value = saved.unit || 'in';
        // Default to true (checked) if not saved
        ignoreLabelHomeEl.checked = saved.ignoreLabelHome !== false;

        // Set values based on saved unit
        const dpi = parseInt(dpiEl.value, 10);
        if (saved.unit === 'dots') {
            widthEl.value = saved.width;
            heightEl.value = saved.height;
            widthEl.dataset.dots = saved.width;
            heightEl.dataset.dots = saved.height;
        } else {
            // Convert saved values to dots first
            const widthDots = displayToDots(saved.width, dpi, saved.unit);
            const heightDots = displayToDots(saved.height, dpi, saved.unit);
            widthEl.dataset.dots = widthDots;
            heightEl.dataset.dots = heightDots;
            widthEl.value = saved.width;
            heightEl.value = saved.height;
        }
    } else {
        // Defaults: 4x6 inches at 203 DPI
        dpiEl.value = DEFAULT_DPI;
        unitEl.value = 'in';
        widthEl.value = DEFAULT_WIDTH_INCHES;
        heightEl.value = DEFAULT_HEIGHT_INCHES;
        widthEl.dataset.dots = DEFAULT_WIDTH_INCHES * DEFAULT_DPI;
        heightEl.dataset.dots = DEFAULT_HEIGHT_INCHES * DEFAULT_DPI;
        ignoreLabelHomeEl.checked = true; // Default to ignoring label home offsets
    }

    updateDimensionDisplays();
}

// Control change handlers
document.getElementById('dpi').addEventListener('change', () => {
    // When DPI changes, update the dots values but keep the display unit values
    const dpi = parseInt(document.getElementById('dpi').value, 10);
    const unit = document.getElementById('unit').value;
    const widthEl = document.getElementById('width');
    const heightEl = document.getElementById('height');

    // Recalculate dots from current display values
    widthEl.dataset.dots = displayToDots(widthEl.value, dpi, unit);
    heightEl.dataset.dots = displayToDots(heightEl.value, dpi, unit);

    render();
    saveState();
});

document.getElementById('unit').addEventListener('change', () => {
    updateDimensionDisplays();
    saveState();
});

document.getElementById('width').addEventListener('input', () => {
    const dpi = parseInt(document.getElementById('dpi').value, 10);
    const unit = document.getElementById('unit').value;
    const widthEl = document.getElementById('width');
    widthEl.dataset.dots = displayToDots(widthEl.value, dpi, unit);

    clearTimeout(renderTimeout);
    renderTimeout = setTimeout(() => {
        render();
        saveState();
    }, 300);
});

document.getElementById('height').addEventListener('input', () => {
    const dpi = parseInt(document.getElementById('dpi').value, 10);
    const unit = document.getElementById('unit').value;
    const heightEl = document.getElementById('height');
    heightEl.dataset.dots = displayToDots(heightEl.value, dpi, unit);

    clearTimeout(renderTimeout);
    renderTimeout = setTimeout(() => {
        render();
        saveState();
    }, 300);
});

document.getElementById('ignore-label-home').addEventListener('change', () => {
    render();
    saveState();
});

// Drag and drop handling
const dropZone = document.getElementById('drop-zone');
const editorContainer = document.getElementById('editor');

function handleDragOver(e) {
    e.preventDefault();
    e.stopPropagation();
    dropZone.classList.add('drag-over');
}

function handleDragLeave(e) {
    e.preventDefault();
    e.stopPropagation();
    dropZone.classList.remove('drag-over');
}

function handleDrop(e) {
    e.preventDefault();
    e.stopPropagation();
    dropZone.classList.remove('drag-over');

    const files = e.dataTransfer.files;
    if (files.length > 0) {
        const file = files[0];
        const reader = new FileReader();
        reader.onload = function (event) {
            if (editor) {
                editor.setValue(event.target.result);
                saveState();
            }
        };
        reader.readAsText(file);
    }
}

dropZone.addEventListener('dragover', handleDragOver);
dropZone.addEventListener('dragleave', handleDragLeave);
dropZone.addEventListener('drop', handleDrop);
editorContainer.addEventListener('dragover', handleDragOver);
editorContainer.addEventListener('dragleave', handleDragLeave);
editorContainer.addEventListener('drop', handleDrop);

// Prevent default drag behavior on document
document.addEventListener('dragover', (e) => e.preventDefault());
document.addEventListener('drop', (e) => e.preventDefault());

// Initialize
initControls();
initWasm();
