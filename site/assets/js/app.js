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

// Example ZPL templates with size presets (width, height in inches)
const examples = {
    hello: {
        width: 4,
        height: 3,
        zpl: `^XA
^FO50,50
^A0N,30,30
^FDHello, World!^FS

^FO50,100
^BQN,2,5
^FDMM,Ahttps://github.com/StirlingMarketingGroup/go-zpl^FS

^FO250,50
^A0N,25,25
^FDgo-zpl WASM Demo^FS

^XZ`
    },

    shipping: {
        width: 4,
        height: 6,
        zpl: `^XA

^FX Top section with logo, name and address.
^CF0,60
^FO50,50^GB100,100,100^FS
^FO75,75^FR^GB100,100,100^FS
^FO93,93^GB40,40,40^FS
^FO220,50^FDIntershipping, Inc.^FS
^CF0,30
^FO220,115^FD1000 Shipping Lane^FS
^FO220,155^FDShelbyville TN 38102^FS
^FO220,195^FDUnited States (USA)^FS
^FO50,250^GB700,3,3^FS

^FX Second section with recipient address and permit information.
^CFA,30
^FO50,300^FDJohn Doe^FS
^FO50,340^FD100 Main Street^FS
^FO50,380^FDSpringfield TN 39021^FS
^FO50,420^FDUnited States (USA)^FS
^CFA,15
^FO600,300^GB150,150,3^FS
^FO638,340^FDPermit^FS
^FO638,390^FD123456^FS
^FO50,500^GB700,3,3^FS

^FX Third section with bar code.
^BY5,2,270
^FO100,550^BC^FD12345678^FS

^FX Fourth section (the two boxes on the bottom).
^FO50,900^GB700,250,3^FS
^FO400,900^GB3,250,3^FS
^CF0,40
^FO100,960^FDCtr. X34B-1^FS
^FO100,1010^FDREF1 F00B47^FS
^FO100,1060^FDREF2 BL4H8^FS
^CF0,190
^FO470,955^FDCA^FS

^XZ`
    },

    ups: {
        width: 4,
        height: 6,
        zpl: `^XA
^LH10,12
^PW812
^CI27
^FO284,524^BY3^BCN,107,N,N,N,A^FD12345678901^FS
^FO66,792^BY3^BCN,208,N,N,N,A^FD1Z00A00A0000000001^FS
^FO20,431^BD2^FD[)>_1E01_1D961Z00000001_1DUPSN_1D00A00A_1E_04^FS
^FO15,7^A0N,20,24^FDJOHN DOE^FS
^FO15,27^A0N,20,24^FD15551234567^FS
^FO15,47^A0N,20,24^FDTEST COMPANY INC^FS
^FO15,67^A0N,20,24^FD123 SENDER STREET, SUITE 100^FS
^FO15,87^A0N,20,24^FDTEST CITY   12345^FS
^FO15,108^A0N,20,24^FDTEST COUNTRY^FS
^FO15,142^A0N,28,32^FDSHIP TO: ^FS
^FO61,166^A0N,28,32^FDJANE SMITH^FS
^FO61,194^A0N,28,32^FD15559876543^FS
^FO61,222^A0N,28,32^FDACME LOGISTICS^FS
^FO61,251^A0N,28,32^FD456 RECEIVER ROAD^FS
^FO61,279^A0N,45,44^FDANYTOWN  CA  90210^FS
^FO61,324^A0N,45,44^FDUNITED STATES^FS
^FO446,9^A0N,30,34^FD6 LBS^FS
^FO683,9^A0N,28,32^FD1 OF 1^FS
^FO508,51^A0N,22,26^FDSHP#: 00A0 0AA0 AA0^FS
^FO508,73^A0N,22,26^FDSHP WT: 6 LBS^FS
^FO508,95^A0N,22,26^FDDATE: 01 JAN 2025^FS
^FO508,117^A0N,22,26^FDDWT: 9,8,7^FS
^FO269,436^A0N,80,70^FDCA 902 1-00^FS
^FO10,1031^A0N,22,26^FDBILLING: F/C RECEIVER 00A00A^FS
^FO10,1053^A0N,22,26^FDDESC: Test merchandise for testing^FS
^FO10,1075^A0N,22,26^FDIMPORT CONTROL - PAYMENT GUARANTEED^FS
^FO543,1035^A0N,60,64^FDEDI-RS^FS
^FO10,1163^A0N,22,26^FDReference No.1: 000001^FS
^FO180,1203^A0N,14,20^FDXOL 01.01.25          NV00 0.0A 01/2025*^FS
^FO676,674^A0N,100,76^FD2^FS
^FO9,670^A0N,56,58^FDUPS EXPEDITED^FS
^FO9,731^A0N,26,30^FDTRACKING #: 1Z 00A 00A 00 0000 0001^FS
^FO0,648^GB811,14,14,B,0^FS
^FO0,423^GB812,4,4,B,0^FS
^FO244,423^GB4,225,4,B,0^FS
^FO0,774^GB812,5,5,B,0^FS
^FO0,1013^GB812,14,14,B,0^FS
^FO629,1147
^GFA,00969,00969,019,FFFFFFFFFFFFFFFFFFFFFFFFFFFFF000000000
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF000000000
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF000000000
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF000000000
F0000000000001F8000000000000F000000000
F0000000000001F8000000000000F000000000
F0000000003F81F83FC000000000F000000000
F0000000003F81F83FC000000000F000000000
F000000000FFF9F9FFF000000000F000000000
F000000000FFF9F9FFF000000000F000000000
F000000000FFFFFFFFFC00000000F000000000
F000000000FFFFFFFFFC00000000F000000000
F000000000F07FFFF0FC00000000F000000000
F000000000F07FFFF0FC00000000F000000000
F000000000FC1FFFC3F000000000F000000000
F000000000FC1FFFC3F000000000F000000000
F000000000FFFFFFFFF000000000F000000000
F000000000FFFFFFFFF000000000F000000000
F0000000003FFFFFFFC000000000F000000000
F0000000003FFFFFFFC000000000F000000000
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF000000000
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF000000000
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF000000000
F00000000001FFFFF00000000000F000000000
F00000000001FFFFF00000000000F000000000
F00000000003FFF9FC0000000000F000000000
F00000000003FFF9FC0000000000F000000000
F0000000003FE1F87FC000000000F000000000
F0000000003FE1F87FC000000000F000000000
F000000000FF81F83FF000000000F000000000
F000000000FF81F83FF000000000F000000000
F000000000FE01F803F000000000F000000000
F000000000FE01F803F000000000F000000000
F000000000F001F800F000000000F000000000
F000000000F001F800F000000000F000000000
F0000000000001F8000000000000F000000000
F0000000000001F8000000000000F0FFDC1C00
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF0FFDC1C00
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF00C1E3C00
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF00C1E3C00
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF00C1A2C00
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF00C1B6C00
FFFFFFFFFFFFFFFFFFFFFFFFFFFFF00C1B6C00
0000000000000000000000000000000C19CC00
0000000000000000000000000000000C19CC00
0000000000000000000000000000000C19CC00
0000000000000000000000000000000C188C00
00000000000000000000000000000000000000
00000000000000000000000000000000000000
00000000000000000000000000000000000000
^FS
^XZ`
    },

    product: {
        width: 4,
        height: 3,
        zpl: `^XA

^FX Product Label with Price and Barcode
^CF0,60
^FO50,30^FDWireless Mouse^FS

^CF0,30
^FO50,100^FDSKU: WM-2024-BLK^FS
^FO50,140^FDColor: Matte Black^FS
^FO50,180^FDBluetooth 5.0 | 2.4GHz^FS

^FX Price box
^FO500,30^GB280,120,3^FS
^CF0,80
^FO530,50^FD$29.99^FS

^FX Horizontal divider
^FO50,220^GB730,3,3^FS

^FX Product barcode
^BY3,2,100
^FO200,260^BC^FD012345678905^FS

^FX QR code for product page
^FO50,400^BQN,2,4^FDMM,Ahttps://example.com/product/wm-2024^FS

^FX Product details
^CF0,25
^FO220,420^FDScan for specs,^FS
^FO220,450^FDreviews & manual^FS

^FX Bottom info
^FO50,520^GB730,3,3^FS
^CF0,20
^FO50,540^FDMade in USA | 1 Year Warranty | RoHS Compliant^FS

^XZ`
    },

    qrcode: {
        width: 4,
        height: 3.5,
        zpl: `^XA

^FX Business Card Style QR Code
^CF0,50
^FO100,50^FDJane Smith^FS
^CF0,30
^FO100,110^FDSenior Developer^FS
^FO100,150^FDAcme Technologies^FS

^FX Divider line
^FO100,200^GB600,3,3^FS

^FX Contact QR Code (vCard format)
^FO100,230^BQN,2,8^FDMM,ABEGIN:VCARD
VERSION:3.0
N:Smith;Jane
FN:Jane Smith
ORG:Acme Technologies
TITLE:Senior Developer
TEL:+1-555-123-4567
EMAIL:jane.smith@acme.dev
URL:https://acme.dev
END:VCARD^FS

^FX Contact details
^CF0,25
^FO480,250^FD+1 (555) 123-4567^FS
^FO480,290^FDjane.smith@acme.dev^FS
^FO480,330^FDacme.dev^FS

^FX GitHub
^FO480,390^FDgithub.com/janesmith^FS

^FX Scan prompt
^CF0,20
^FO200,600^FDScan to add contact^FS

^XZ`
    },

    shapes: {
        width: 4,
        height: 3,
        zpl: `^XA

^FX Shapes Demonstration
^CF0,40
^FO300,20^FDShapes Demo^FS

^FX Boxes with different thicknesses
^FO50,80^GB100,100,2^FS
^FO170,80^GB100,100,5^FS
^FO290,80^GB100,100,10^FS
^FO410,80^GB100,100,50^FS

^CF0,20
^FO65,190^FDThin^FS
^FO175,190^FDMedium^FS
^FO295,190^FDThick^FS
^FO425,190^FDFilled^FS

^FX Circles
^FO50,230^GC80,2^FS
^FO150,230^GC80,5^FS
^FO250,230^GC80,10^FS
^FO350,230^GC80,40^FS

^FX Ellipses
^FO50,340^GE150,60,3^FS
^FO220,340^GE100,80,5^FS
^FO340,340^GE80,100,8^FS

^FX Diagonal lines
^FO50,450^GD100,100,5,B,R^FS
^FO170,450^GD100,100,5,B,L^FS
^FO290,450^GD150,80,8,B,R^FS

^FX Field Reverse Demo
^FO500,250^GB250,150,150^FS
^FO520,270^FR^GB210,110,110^FS
^FO550,300^GB150,50,50^FS

^CF0,18
^FO530,420^FDNested Boxes^FS
^FO530,440^FD(Field Reverse)^FS

^XZ`
    },

    barcodes: {
        width: 4,
        height: 3,
        zpl: `^XA

^FX Barcode Sampler
^CF0,35
^FO250,20^FDBarcode Sampler^FS

^FX Code 128
^CF0,20
^FO50,70^FDCode 128:^FS
^BY2,2,60
^FO50,95^BC^FDCODE128-TEST^FS

^FX QR Code
^FO450,70^FDQR Code:^FS
^FO450,95^BQN,2,4^FDMM,Ahttps://go-zpl.dev^FS

^FX Code 39
^FO50,200^FDCode 39:^FS
^BY2,2,60
^FO50,225^B3N,N,60,Y,N^FDCODE39^FS

^FX DataMatrix
^FO450,200^FDDataMatrix:^FS
^FO450,225^BXN,5,200^FDDataMatrix Test 123^FS

^FX Interleaved 2 of 5
^FO50,330^FDInterleaved 2 of 5:^FS
^BY2,2,60
^FO50,355^B2N,60,Y,N,N^FD12345678^FS

^FX PDF417
^FO450,330^FDPDF417:^FS
^FO450,355^B7N,5,3,2,5,N^FDPDF417 Demo^FS

^FX UPC-A
^FO50,460^FDUPC-A:^FS
^BY2,2,60
^FO50,485^BUN,60,Y,N^FD01234567890^FS

^FX MaxiCode (if supported)
^FO450,460^FDMaxiCode:^FS
^FO450,485^BD2^FD[)>_1E01_1D961Z00000001_1DUPSN_1D12345_1E_04^FS

^XZ`
    },
};

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
document.getElementById('example').addEventListener('change', () => {
    const exampleKey = document.getElementById('example').value;
    if (exampleKey && examples[exampleKey] && editor) {
        const example = examples[exampleKey];

        // Set the ZPL content
        editor.setValue(example.zpl);

        // Apply size presets if available
        if (example.width && example.height) {
            const dpi = parseInt(document.getElementById('dpi').value, 10);
            const unitEl = document.getElementById('unit');
            const widthEl = document.getElementById('width');
            const heightEl = document.getElementById('height');

            // Store dots values
            widthEl.dataset.dots = Math.round(example.width * dpi);
            heightEl.dataset.dots = Math.round(example.height * dpi);

            // Set to inches and update display
            unitEl.value = 'in';
            widthEl.value = example.width;
            heightEl.value = example.height;
            widthEl.dataset.unit = 'in';
            heightEl.dataset.unit = 'in';

            updateDimensionDisplays();
        }

        saveState();
    }
});

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
