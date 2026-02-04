// Quick test for libzpl
// From cmd/libzpl directory:
//   Build library: go build -buildmode=c-shared -o libzpl.so ./
//   Compile test:  gcc -o testdata/test_libzpl testdata/test.c -L. -lzpl -I.
//   Run:           cd testdata && LD_LIBRARY_PATH=.. ./test_libzpl

#include <stdio.h>
#include <string.h>
#include "../libzpl.h"

int main() {
    const char* zpl = "^XA^FO50,50^A0N,30,30^FDHello from C!^FS^XZ";
    char* png_data = NULL;
    int png_len = 0;

    printf("Testing zpl_render_png_simple...\n");

    int result = zpl_render_png_simple(
        (char*)zpl,
        strlen(zpl),
        &png_data,
        &png_len
    );

    if (result == 0) {
        printf("Success! PNG size: %d bytes\n", png_len);

        // Verify it looks like a PNG (magic bytes)
        if (png_len > 8 &&
            (unsigned char)png_data[0] == 0x89 &&
            png_data[1] == 'P' &&
            png_data[2] == 'N' &&
            png_data[3] == 'G') {
            printf("PNG header verified!\n");
        }

        // Write to file for visual inspection
        FILE* f = fopen("test_output.png", "wb");
        if (f) {
            fwrite(png_data, 1, png_len, f);
            fclose(f);
            printf("Wrote test_output.png\n");
        }

        zpl_free(png_data);
    } else {
        printf("Error: %d\n", result);
        return 1;
    }

    return 0;
}
