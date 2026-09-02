import { Dazzle, DazzleError } from 'dazzle-zpl';

// ZPL Label Renderer - Client-side application

let editor;
let renderTimeout;
let wasmReady = false;

// Base64 for the WASM boundary and saved state: the same bytes Dazzle prints.
function safeBase64Encode(str) {
    return btoa(bytesToLatin1(zplBytes(str)));
}

// Inverse of zplBytes for imported bytes. Not TextDecoder('iso-8859-1'): browsers
// alias that label to windows-1252, which remaps 0x80–0x9F (0x80 → U+20AC) and
// breaks the 1:1 byte mapping binary ^GF data depends on.
function bytesToLatin1(bytes) {
    let out = '';
    for (let i = 0; i < bytes.length; i += 0x8000) {
        out += String.fromCharCode.apply(null, bytes.subarray(i, i + 0x8000));
    }
    return out;
}

// The editor document as bytes: an imported file is a Latin-1 byte string (1:1 with
// the file, so binary ^GF data survives), and only genuinely Unicode text typed into
// the editor gets UTF-8-encoded.
function zplBytes(str) {
    if (/^[\u0000-\u00ff]*$/.test(str)) {
        return Uint8Array.from(str, (c) => c.charCodeAt(0));
    }
    return new TextEncoder().encode(str);
}

// Storage keys
const STORAGE_KEY = 'zpl-renderer';
// The Dazzle opt-in is its own key, not a field of the editor state: saveState()
// serializes the editor document, so writing the opt-in through it before Monaco
// has loaded would overwrite the saved label with the default.
const DAZZLE_STORAGE_KEY = 'zpl-renderer-dazzle';

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

    labelary: {
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
^FO20,431^BD2^FH_^FD000000000000000[)>_1E01_1D961Z00000001_1DUPSN_1D00A00A_1E07Y+0*0A.AA'AA#A0A%'_0DAAA0.00_1C*0AAA'A_1C0AA000$&A_0D_1E_04^FS
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
^BY3,2,80
^FO220,250^BC^FD012345678905^FS

^FX QR code for product page
^FO50,380^BQN,2,3^FDMM,Ahttps://example.com/product/wm-2024^FS

^FX Product details
^CF0,25
^FO180,395^FDScan for specs,^FS
^FO180,425^FDreviews & manual^FS

^FX Bottom info
^FO50,500^GB730,3,3^FS
^CF0,20
^FO50,520^FDMade in USA | 1 Year Warranty | RoHS Compliant^FS

^XZ`
    },

    qrcode: {
        width: 4,
        height: 3,
        zpl: `^XA

^FX Business Card Style QR Code
^CF0,45
^FO50,40^FDJane Smith^FS
^CF0,28
^FO50,95^FDSenior Developer^FS
^FO50,130^FDAcme Technologies^FS

^FX Divider line
^FO50,175^GB710,3,3^FS

^FX Contact QR Code (vCard format)
^FO50,200^BQN,2,4^FDMM,ABEGIN:VCARD
VERSION:3.0
N:Smith;Jane
FN:Jane Smith
ORG:Acme Technologies
TITLE:Senior Developer
TEL:+1-555-123-4567
EMAIL:jane.smith@acme.dev
URL:https://acme.dev
END:VCARD^FS

^FX Contact details (to the right of QR)
^CF0,24
^FO300,220^FD+1 (555) 123-4567^FS
^FO300,255^FDjane.smith@acme.dev^FS
^FO300,290^FDacme.dev^FS
^FO300,325^FDgithub.com/janesmith^FS

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

    fedexGround: {
        width: 4,
        height: 6,
        zpl: `^XA^CF,0,0,0^PR12^MD30^PW800^POI^CI13^LH0,20
^FO12,124^GB755,2,2^FS
^FO12,390^GB777,2,2^FS
^FO32,3^AdN,0,0^FWN^FH^FDFROM:^FS
^FO32,19^AdN,0,0^FWN^FH^FDTest Sender^FS
^FO32,37^AdN,0,0^FWN^FH^FDTest Company Inc^FS
^FO32,55^AdN,0,0^FWN^FH^FD123 Test Street^FS
^FO32,73^AdN,0,0^FWN^FH^FD^FS
^FO32,109^AdN,0,0^FWN^FH^FDUS ^FS
^FO224,3^AdN,0,0^FWN^FH^FD(555) 123-4567^FS
^FO28,742^A0N,24,24^FWN^FH^FDTRK#^FS
^FO28,800^A0N,27,32^FWN^FH^FD^FS
^FO136,712^A0N,27,36^FWN^FH^FD^FS
^FO32,91^AdN,0,0^FWN^FH^FDMemphis TN 38118^FS
^FO478,3^AdN,0,0^FWN^FH^FDSHIP DATE: 30JAN26^FS
^FO478,19^AdN,0,0^FWN^FH^FDACTWGT: 3.00 LB^FS
^FO478,37^AdN,0,0^FWN^FH^FDCAD: 0000000/FAPI2208^FS
^FO478,91^AdN,0,0^FWN^FH^FDBILL SENDER^FS
^FO39,136^A0N,39,39^FWN^FH^FDTest Recipient^FS
^FO39,178^A0N,39,39^FWN^FH^FDRecipient Corp^FS
^FO39,220^A0N,39,39^FWN^FH^FD456 Delivery Ave^FS
^FO39,262^A0N,39,39^FWN^FH^FD**TEST LABEL - DO NOT SHIP**^FS
^FO39,347^AdN,0,0^FWN^FH^FD(555) 987-6543^FS
^FO39,304^A0N,43,40^FWN^FH^FDDallas TX 75201^FS
^FO719,304^A0N,43,40^FWN^FH^FD(US)^FS
^FO709,440^A0N,19,26^FWN^FH^FDGround^FS
^FO689,480^A0N,128,137^FWN^FH^FDG^FS
^FO677,462^GB104,10,10^FS
^FO677,472^GB10,112,10^FS
^FO771,472^GB10,112,10^FS
^FO677,584^GB104,10,10^FS
^FO654,402^A0N,43,58^FWN^FH^FDFedEx^FS
^FO709,440^A0N,19,26^FWN^FH^FDGround^FS
^FO689,480^A0N,128,137^FWN^FH^FDG^FS
^FO791,493^A0N,13,18^FWB^FH^FDJ261026012001uv^FS
^FO9,136^A0N,21,21^FWN^FH^FDTO^FS
^FO21,412^BY2,2^B7N,10,5,14^FH^FWN^FH^FD[)>_1E01_1D0275201_1D840_1D019_1D794981365794_1DFDEG_1D4910221_1D030_1D_1D1/1_1D3.00LB_1DN_1D456 Delivery Ave_1DDallas_1DTX_1DTest Recipient_1E06_1D10ZGD009_1D11ZRecipient Corp_1D12Z5559876543_1D20Z_1C_1D31Z9622001900004910221300794981365794_1D34Z01_1D_1E_04^FS
^FO28,837^A0N,107,96^FWN^FH^FD^FS
^FO12,681^GB777,2,2^FS
^FO494,885^A0N,43,43^FWN^FH^FD^FS
^FO788,28^AbN,11,7^FWB^FH^FD58KJ4/747B/484B^FS
^FO95,746^A0N,53,40^FWN^FH^FD0000 0000 0000^FS
^FO409,695^A0N,51,38^FWN^FH^FB390,,,R,^FD                   ^FS
^FO404,747^A0N,51,38^FWN^FH^FB400,,,R,^FD                   ^FS
^FO413,799^A0N,40,40^FWN^FH^FB386,,,R,^FD                ^FS
^FO495,841^A0N,44,44^FWN^FH^FB298,,,R,^FD     75201^FS
^FO574,901^A0N,24,24^FWN^FH^FB120,,,R,^FD      ^FS
^FO695,885^A0N,43,43^FWN^FH^FB100,,,R,^FD   ^FS
^FO39,927^A0N,27,36^FWN^FH^FD0000 0000 0 (000 000 0000) 0 00 0000 0000 0000^FS
^FO75,968^BY3,2^BCN,200,N,N,N,N^FWN^FD>;9622001900004910221300000000000000^FS
^FO135,1028^A0N,128,137^FWN^FH^FDSAMPLE^FS
^FO478,55^AdN,0,0^FWN^FH^FDDIMMED: 10 X 8 X 4 IN^FS
^FO329,349^AbN,11,7^FWN^FH^FDREF: ^FS
^FO39,363^AbN,11,7^FWN^FH^FDINV: ^FS
^FO39,377^AbN,11,7^FWN^FH^FDPO: ^FS
^FO429,377^AbN,11,7^FWN^FH^FDDEPT: ^FS
^PQ1
^XZ`
    },

    fedexExpress: {
        width: 4,
        height: 6,
        zpl: `^XA^CF,0,0,0^PR12^MD30^PW800^POI^CI13^LH0,20
^FO12,139^GB753,2,2^FS
^FO12,405^GB777,2,2^FS
^FO464,8^GB2,129,2^FS
^FO28,747^A0N,24,24^FWN^FH^FDTRK#^FS
^FO28,805^A0N,27,32^FWN^FH^FD^FS
^FO136,717^A0N,27,36^FWN^FH^FD^FS
^FO32,10^AdN,0,0^FWN^FH^FDORIGIN ID:HKAA ^FS
^FO224,10^AdN,0,0^FWN^FH^FD(901) 555-1234^FS
^FO32,28^AdN,0,0^FWN^FH^FDTEST SENDER^FS
^FO32,46^AdN,0,0^FWN^FH^FDSMG TEST COMPANY^FS
^FO32,64^AdN,0,0^FWN^FH^FD3620 HACKS CROSS ROAD^FS
^FO32,82^AdN,0,0^FWN^FH^FD^FS
^FO32,100^AdN,0,0^FWN^FH^FDMEMPHIS, TN 38125^FS
^FO32,118^AdN,0,0^FWN^FH^FDUNITED STATES US^FS
^FO478,46^AdN,0,0^FWN^FH^FDCAD: 0000000/FAPI2208^FS
^FO15,151^A0N,21,21^FWN^FH^FDTO^FS
^FO60,149^A0N,38,38^FWN^FH^FDTEST RECIPIENT^FS
^FO60,191^A0N,38,38^FWN^FH^FDRECIPIENT CORP^FS
^FO60,233^A0N,38,38^FWN^FH^FD100 MARKET STREET^FS
^FO60,275^A0N,38,38^FWN^FH^FD**TEST LABEL - DO NOT SHIP**^FS
^FO60,317^A0N,43,40^FWN^FH^FDSAN FRANCISCO CA 94105^FS
^FO35,359^A0N,21,21^FWN^FH^FD(415) 555-9876^FS
^FO677,511^GB104,10,10^FS
^FO677,521^GB10,112,10^FS
^FO771,521^GB10,112,10^FS
^FO677,633^GB104,10,10^FS
^FO652,449^A0N,43,58^FWN^FH^FDFedEx^FS
^FO708,488^A0N,19,26^FWN^FH^FDExpress^FS
^FO697,529^A0N,128,137^FWN^FH^FDE^FS
^FO21,416^BY2,3^BCN,25,N,N,N^FWN^FD7100 MARKET STREET^FS
^FO21,449^BY2,2^B7N,10,5,14^FH^FWN^FH^FD[)>_1E01_1D0294105_1D840_1D20_1D7949819308110201_1DFDE_1D740561073_1D031_1D_1D1/1_1D5.00LB_1DN_1D100 Market Street_1DSan Francisco_1DCA_1DTest Recipient_1E06_1D10ZED008_1D11ZRecipient Corp_1D12Z4155559876_1D15Z114064860_1D20Z_1C_1D31Z1195282044690009410500794981930811_1D32Z02GD_1D34Z01_1D39ZHKAA_1D_1E09_1DFDX_1Dz_1D8_1D_17_04';0?_7F@_1E_04^FS
^FO478,100^AdN,0,0^FWN^FH^FDBILL SENDER^FS
^FO12,694^GB777,2,2^FS
^FO494,890^A0N,43,43^FWN^FH^FD^FS
^FO791,120^AbN,11,7^FWB^FH^FD58KJ4/747B/484B^FS
^FO95,751^A0N,53,40^FWN^FH^FD0000 0000 0000^FS
^FO409,700^A0N,51,38^FWN^FH^FB390,,,R,^FD WED - 04 FEB 5:00P^FS
^FO309,752^A0N,51,38^FWN^FH^FB490,,,R,^FD      EXPRESS SAVER^FS
^FO413,804^A0N,40,40^FWN^FH^FB386,,,R,^FD                ^FS
^FO495,846^A0N,44,44^FWN^FH^FB298,,,R,^FDSFOA 94105^FS
^FO574,906^A0N,24,24^FWN^FH^FB120,,,R,^FD CA-US^FS
^FO695,890^A0N,43,43^FWN^FH^FB100,,,R,^FDSFO^FS
^FO39,932^A0N,27,32^FWN^FH^FD^FS
^FO75,993^BY3,2^BCN,200,N,N,N,N^FWN^FD>;1195282044690009410500000000000000^FS
^FO135,1053^A0N,128,137^FWN^FH^FDSAMPLE^FS
^FO28,842^A0N,107,96^FWN^FH^FDSS SFOAG^FS
^FO790,523^A0N,13,18^FWB^FH^FDJ261026012001uv^FS
^FO478,10^AdN,0,0^FWN^FH^FDSHIP DATE: 31JAN26^FS
^FO478,28^AdN,0,0^FWN^FH^FDACTWGT: 5.00 LB^FS
^FO478,64^AdN,0,0^FWN^FH^FDDIMS: 12x10x6 IN^FS
^FO328,364^AbN,11,7^FWN^FH^FDREF: ^FS
^FO38,378^AbN,11,7^FWN^FH^FDINV: ^FS
^FO38,392^AbN,11,7^FWN^FH^FDPO: ^FS
^FO428,392^AbN,11,7^FWN^FH^FDDEPT: ^FS
^FO25,768^GB58,1,1^FS
^FO25,768^GB1,26,1^FS
^FO83,768^GB1,26,1^FS
^FO25,794^GB58,1,1^FS
^FO31,774^AdN,0,0^FWN^FH^FD0201^FS
^PQ1
^XZ`
    },

    barcodes: {
        width: 4,
        height: 5,
        zpl: `^XA

^FX Barcode Sampler - shows supported barcode types
^CF0,35
^FO250,20^FDBarcode Sampler^FS

^FX Code 128
^CF0,20
^FO50,70^FDCode 128:^FS
^BY2,2,80
^FO50,95^BC^FDCODE128-TEST^FS

^FX QR Code
^FO500,70^FDQR Code:^FS
^FO500,95^BQN,2,4^FDMM,Ahttps://go-zpl.dev^FS

^FX DataMatrix
^FO50,250^FDDataMatrix:^FS
^FO50,275^BXN,6,200^FDDataMatrix 123^FS

^FX MaxiCode
^FO450,250^FDMaxiCode:^FS
^FO420,275^BD2^FH_^FD000000000000000[)>_1E01_1D961Z00000001_1DUPSN_1D12345_1E07TESTDATA_0D_1E_04^FS

^FX PDF417 - on its own row since it's wide
^FO50,480^FDPDF417:^FS
^FO50,505^B7N,50,2^FDPDF417 Demo^FS

^FX Note about other barcodes
^CF0,16
^FO50,620^FDSupported: Code 128, QR, DataMatrix, PDF417, MaxiCode^FS
^FO50,645^FDComing soon: Code 39, UPC-A, EAN-13, Interleaved 2of5^FS

^XZ`
    },

    uspsDomestic: {
        width: 4,
        height: 6,
        file: 'examples/usps_domestic.zpl.b64',
        isBase64: true
    },

    uspsApo: {
        width: 4,
        height: 6,
        file: 'examples/usps_apo.zpl.b64',
        isBase64: true
    },

    uspsIntl: {
        width: 4,
        height: 6,
        file: 'examples/usps_intl.zpl.b64',
        isBase64: true
    },

    uspsApoContinuation: {
        width: 4,
        height: 6,
        file: 'examples/usps_apo_continuation.zpl.b64',
        isBase64: true
    },

    fedexIntlExpressAwb: {
        width: 4,
        height: 6,
        file: 'examples/fedex_intl_express_awb.zpl.b64',
        isBase64: true
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
        const zplContent = editor ? editor.getValue() : defaultZPL;
        const state = {
            dpi: document.getElementById('dpi').value,
            width: document.getElementById('width').value,
            height: document.getElementById('height').value,
            unit: document.getElementById('unit').value,
            ignoreLabelHome: document.getElementById('ignore-label-home').checked,
            // Store ZPL as base64 to preserve binary data
            zplBase64: safeBase64Encode(zplContent),
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
window.require.config({ paths: { vs: 'https://cdn.jsdelivr.net/npm/monaco-editor@0.45.0/min/vs' } });

window.require(['vs/editor/editor.main'], function () {
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

    // Load saved state (default to UPS Import Control example)
    const saved = loadState();
    // Decode base64 if present, fall back to old format or default
    let initialZPL;
    if (saved?.zplBase64) {
        initialZPL = atob(saved.zplBase64);
    } else if (saved?.zpl) {
        initialZPL = saved.zpl; // Legacy format
    } else {
        initialZPL = examples.ups.zpl;
    }

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
    const previewContainer = document.getElementById('preview-container');
    const renderTimeEl = document.getElementById('render-time');

    const startTime = performance.now();

    try {
        const ignoreLabelHome = document.getElementById('ignore-label-home').checked;
        // Always encode as base64 for WASM (preserves binary data through JS->Go transition)
        // Use safeBase64Encode to handle any characters that Monaco might introduce
        const base64Data = safeBase64Encode(zpl);
        const result = window.renderZPL(base64Data, dpi, dims.width, dims.height, ignoreLabelHome, true);
        const elapsed = (performance.now() - startTime).toFixed(1);

        if (result.error) {
            errorEl.textContent = result.error;
            errorEl.style.display = 'block';
            previewContainer.style.display = 'none';
            renderTimeEl.textContent = '';
        } else {
            errorEl.style.display = 'none';

            // Clear previous images
            previewContainer.innerHTML = '';

            // Handle multiple pages
            const images = result.images || [result.image];

            images.forEach((imgData, index) => {
                const imgSrc = 'data:image/png;base64,' + imgData;

                // Wrap in Fancybox link for zoom
                const link = document.createElement('a');
                link.href = imgSrc;
                link.className = 'preview-link';
                link.setAttribute('data-fancybox', 'labels');
                link.setAttribute('data-caption', `Page ${index + 1} of ${images.length}`);

                const img = document.createElement('img');
                img.src = imgSrc;
                img.alt = `Page ${index + 1}`;
                img.className = 'preview-image';

                link.appendChild(img);
                previewContainer.appendChild(link);
            });

            previewContainer.style.display = 'grid';

            // Set max-height on images based on container size
            const containerHeight = previewContainer.clientHeight - 32; // subtract padding
            previewContainer.querySelectorAll('.preview-image').forEach(img => {
                img.style.maxHeight = containerHeight + 'px';
            });

            // Rebind Fancybox to new elements
            if (typeof Fancybox !== 'undefined') {
                Fancybox.destroy();
                Fancybox.bind('[data-fancybox]', {
                    Thumbs: { type: 'classic' }
                });
            }
            const pageText = images.length > 1 ? ` (${images.length} pages)` : '';
            renderTimeEl.textContent = `${result.width}x${result.height} rendered in ${elapsed}ms${pageText}`;
        }
    } catch (err) {
        errorEl.textContent = 'Render error: ' + err.message;
        errorEl.style.display = 'block';
        previewContainer.style.display = 'none';
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
document.getElementById('example').addEventListener('change', async () => {
    const exampleKey = document.getElementById('example').value;
    if (exampleKey && examples[exampleKey] && editor) {
        const example = examples[exampleKey];

        // Get ZPL content - either inline or fetch from file
        let zpl;
        if (example.file) {
            try {
                const response = await fetch(example.file);
                if (example.isBase64) {
                    // Fetch base64, decode to binary, convert to Latin-1 string
                    const base64 = await response.text();
                    const binaryString = atob(base64);
                    // binaryString is already a Latin-1 string (each char = one byte)
                    zpl = binaryString;
                } else {
                    const buffer = await response.arrayBuffer();
                    zpl = bytesToLatin1(new Uint8Array(buffer));
                }
            } catch (e) {
                console.error('Failed to fetch example:', e);
                return;
            }
        } else {
            zpl = example.zpl;
        }

        // Set the ZPL content
        editor.setValue(zpl);

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
                editor.setValue(bytesToLatin1(new Uint8Array(event.target.result)));
                saveState();
            }
        };
        reader.readAsArrayBuffer(file);
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

// Print functionality
document.getElementById('print-btn').addEventListener('click', () => {
    const previewContainer = document.getElementById('preview-container');
    const images = previewContainer.querySelectorAll('.preview-image');
    if (images.length === 0 || previewContainer.style.display === 'none' || previewContainer.style.display === '') {
        return;
    }

    // Create a new window for printing
    const printWindow = window.open('', '_blank');
    if (!printWindow) {
        // Fallback to regular print if popup blocked
        window.print();
        return;
    }

    const renderTimeEl = document.getElementById('render-time');
    const dimensions = renderTimeEl.textContent.match(/^(\d+)x(\d+)/);
    const dpi = parseInt(document.getElementById('dpi').value, 10);

    // Calculate print size in inches
    let widthInches = 4;
    let heightInches = 6;
    if (dimensions) {
        widthInches = parseInt(dimensions[1], 10) / dpi;
        heightInches = parseInt(dimensions[2], 10) / dpi;
    }

    // Validate image sources are data URLs to prevent injection
    const validImages = Array.from(images).filter(img =>
        img.src.startsWith('data:image/png;base64,')
    );
    if (validImages.length === 0) return;

    // Build image tags for all pages
    const imageTags = validImages.map(img => `<img src="${img.src}">`).join('\n');

    printWindow.document.write(`
        <!DOCTYPE html>
        <html>
        <head>
            <title>Print Label</title>
            <style>
                @page {
                    size: ${widthInches}in ${heightInches}in;
                    margin: 0;
                }
                body {
                    margin: 0;
                    padding: 0;
                }
                img {
                    width: ${widthInches}in;
                    height: ${heightInches}in;
                    object-fit: contain;
                    display: block;
                    page-break-after: always;
                }
                img:last-child {
                    page-break-after: avoid;
                }
            </style>
        </head>
        <body>
            ${imageTags}
        </body>
        </html>
    `);
    printWindow.document.close();

    // Wait for all images to load before printing
    printWindow.onload = () => {
        printWindow.print();
        printWindow.close();
    };
});

const dazzle = new Dazzle();
const dazzlePrintBtn = document.getElementById('dazzle-print-btn');
const dazzleNotice = document.getElementById('dazzle-notice');
const dazzleConnect = document.getElementById('dazzle-connect');
const dazzleDownload = document.getElementById('dazzle-download');
const dazzleConnecting = document.getElementById('dazzle-connecting');
const toasts = document.getElementById('toasts');

function showToast(text, isError) {
    const toast = document.createElement('div');
    toast.className = isError ? 'toast toast-error' : 'toast';
    const icon = document.createElement('i');
    icon.className = isError ? 'fa-solid fa-circle-exclamation' : 'fa-solid fa-circle-check';
    const message = document.createElement('span');
    message.textContent = text;
    toast.append(icon, message);
    toasts.appendChild(toast);
    setTimeout(() => toast.remove(), 5000);
}

// A Dazzle print is idle → sending → sent | failed; this is the one place that
// renders those states onto the button and the toasts.
function renderDazzlePrint(state, text) {
    dazzlePrintBtn.disabled = state === 'sending';
    if (state === 'sent' || state === 'failed') {
        showToast(text, state === 'failed');
    }
}

// The Dazzle connection is notOptedIn → connecting → connected | unavailable, and
// polling moves it between the last two as the server comes and goes; this is the
// one place that renders it.
function renderDazzleConnection(state) {
    dazzleNotice.hidden = state === 'connected';
    dazzleDownload.hidden = state === 'connecting';
    dazzleConnect.hidden = state !== 'notOptedIn';
    dazzleConnecting.hidden = state !== 'connecting';
    dazzlePrintBtn.hidden = state !== 'connected';
}

function startDazzleWatch() {
    renderDazzleConnection('connecting');
    dazzle.watch((running) => {
        renderDazzleConnection(running ? 'connected' : 'unavailable');
    }, { interval: 2000 });
}

// Polling localhost from a public https page triggers Chrome's Local Network Access
// permission prompt; a site that promises "no servers" must not fire that for every
// visitor on load. Poll only after the user says they have Dazzle, and remember it.
function initDazzle() {
    let optedIn = false;
    try {
        optedIn = localStorage.getItem(DAZZLE_STORAGE_KEY) === '1';
    } catch (e) {
        console.warn('Failed to load Dazzle preference:', e);
    }
    if (optedIn) {
        startDazzleWatch();
    } else {
        renderDazzleConnection('notOptedIn');
    }
}

dazzleConnect.addEventListener('click', (e) => {
    e.preventDefault();
    try {
        localStorage.setItem(DAZZLE_STORAGE_KEY, '1');
    } catch (err) {
        console.warn('Failed to save Dazzle preference:', err);
    }
    startDazzleWatch();
});

dazzlePrintBtn.addEventListener('click', async () => {
    if (!editor) return;
    renderDazzlePrint('sending');
    try {
        const result = await dazzle.print(zplBytes(editor.getValue()));
        renderDazzlePrint('sent', `Sent to printer (job ${result.job_id})`);
    } catch (err) {
        const message = err instanceof DazzleError ? err.message : `Dazzle unreachable: ${err.message}`;
        renderDazzlePrint('failed', message);
    }
});

// Initialize
initControls();
initDazzle();
initWasm();
