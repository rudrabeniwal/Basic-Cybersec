#include <stdlib.h>
#include <stdint.h>
#include <string.h>
#include <stdio.h>
#include <errno.h>

#include <string>
#include <vector>

#include <bfd.h> // GNU Binary File Descriptor library

#include "loader.h" // Custom header for the loader

// Function to open a binary file using BFD and return the BFD handle
static bfd* open_bfd(std::string &fname) {
    static int bfd_inited = 0; // Static variable to ensure BFD is initialized only once
    bfd *bfd_h;

    if (!bfd_inited) {
        bfd_init(); // Initialize the BFD library
        bfd_inited = 1;
    }

    bfd_h = bfd_openr(fname.c_str(), NULL); // Open the binary file
    if (!bfd_h) {
        fprintf(stderr, "failed to open binary '%s' (%s)\n",
                fname.c_str(), bfd_errmsg(bfd_get_error()));
        return NULL;
    }

    if (!bfd_check_format(bfd_h, bfd_object)) { // Check if the file is an executable
        fprintf(stderr, "file '%s' does not look like an executable (%s)\n",
                fname.c_str(), bfd_errmsg(bfd_get_error()));
        return NULL;
    }

    bfd_set_error(bfd_error_no_error); // Reset BFD error status

    if (bfd_get_flavour(bfd_h) == bfd_target_unknown_flavour) { // Check if the file format is recognized
        fprintf(stderr, "unrecognized format for binary '%s' (%s)\n",
                fname.c_str(), bfd_errmsg(bfd_get_error()));
        return NULL;
    }

    return bfd_h;
}

// Function to load symbols from the symbol table using BFD
static int load_symbols_bfd(bfd *bfd_h, Binary *bin) {
    int ret;
    long n, nsyms, i;
    asymbol **bfd_symtab;
    Symbol *sym;

    bfd_symtab = NULL;

    n = bfd_get_symtab_upper_bound(bfd_h); // Get the size of the symbol table
    if (n < 0) {
        fprintf(stderr, "failed to read symtab (%s)\n",
                bfd_errmsg(bfd_get_error()));
        goto fail;
    } else if (n) {
        bfd_symtab = (asymbol**)malloc(n); // Allocate memory for the symbol table
        if (!bfd_symtab) {
            fprintf(stderr, "out of memory\n");
            goto fail;
        }
        nsyms = bfd_canonicalize_symtab(bfd_h, bfd_symtab); // Load symbols into the table
        if (nsyms < 0) {
            fprintf(stderr, "failed to read symtab (%s)\n",
                    bfd_errmsg(bfd_get_error()));
            goto fail;
        }
        for (i = 0; i < nsyms; i++) {
            if (bfd_symtab[i]->flags & BSF_FUNCTION) { // Check if the symbol is a function
                bin->symbols.push_back(Symbol());
                sym = &bin->symbols.back();
                sym->type = Symbol::SYM_TYPE_FUNC; // Set symbol type to function
                sym->name = std::string(bfd_symtab[i]->name); // Get symbol name
                sym->addr = bfd_asymbol_value(bfd_symtab[i]); // Get symbol address
            }
        }
    }

    ret = 0;
    goto cleanup;

fail:
    ret = -1;

cleanup:
    if (bfd_symtab) free(bfd_symtab); // Free allocated memory

    return ret;
}

// Function to load dynamic symbols (from the dynamic symbol table) using BFD
static int load_dynsym_bfd(bfd *bfd_h, Binary *bin) {
    int ret;
    long n, nsyms, i;
    asymbol **bfd_dynsym;
    Symbol *sym;

    bfd_dynsym = NULL;

    n = bfd_get_dynamic_symtab_upper_bound(bfd_h); // Get the size of the dynamic symbol table
    if (n < 0) {
        fprintf(stderr, "failed to read dynamic symtab (%s)\n",
                bfd_errmsg(bfd_get_error()));
        goto fail;
    } else if (n) {
        bfd_dynsym = (asymbol**)malloc(n); // Allocate memory for the dynamic symbol table
        if (!bfd_dynsym) {
            fprintf(stderr, "out of memory\n");
            goto fail;
        }
        nsyms = bfd_canonicalize_dynamic_symtab(bfd_h, bfd_dynsym); // Load dynamic symbols into the table
        if (nsyms < 0) {
            fprintf(stderr, "failed to read dynamic symtab (%s)\n",
                    bfd_errmsg(bfd_get_error()));
            goto fail;
        }
        for (i = 0; i < nsyms; i++) {
            if (bfd_dynsym[i]->flags & BSF_FUNCTION) { // Check if the dynamic symbol is a function
                bin->symbols.push_back(Symbol());
                sym = &bin->symbols.back();
                sym->type = Symbol::SYM_TYPE_FUNC; // Set symbol type to function
                sym->name = std::string(bfd_dynsym[i]->name); // Get symbol name
                sym->addr = bfd_asymbol_value(bfd_dynsym[i]); // Get symbol address
            }
        }
    }

    ret = 0;
    goto cleanup;

fail:
    ret = -1;

cleanup:
    if (bfd_dynsym) free(bfd_dynsym); // Free allocated memory

    return ret;
}

// Function to load sections (e.g., code, data) from the binary file using BFD
static int load_sections_bfd(bfd *bfd_h, Binary *bin) {
    int bfd_flags;
    uint64_t vma, size;
    const char *secname;
    asection* bfd_sec;
    Section *sec;
    Section::SectionType sectype;

    for (bfd_sec = bfd_h->sections; bfd_sec; bfd_sec = bfd_sec->next) { // Iterate over all sections
        bfd_flags = bfd_section_flags(bfd_sec); // Get section flags

        sectype = Section::SEC_TYPE_NONE;
        if (bfd_flags & SEC_CODE) {
            sectype = Section::SEC_TYPE_CODE; // Identify code sections
        } else if (bfd_flags & SEC_DATA) {
            sectype = Section::SEC_TYPE_DATA; // Identify data sections
        } else {
            continue;
        }

        vma = bfd_section_vma(bfd_sec); // Get section's virtual memory address
        size = bfd_section_size(bfd_sec); // Get section size
        secname = bfd_section_name(bfd_sec); // Get section name
        if (!secname) secname = "<unnamed>";

        bin->sections.push_back(Section());
        sec = &bin->sections.back();

        sec->binary = bin;
        sec->name = std::string(secname);
        sec->type = sectype;
        sec->vma = vma;
        sec->size = size;
        sec->bytes = (uint8_t*)malloc(size); // Allocate memory for section contents
        if (!sec->bytes) {
            fprintf(stderr, "out of memory\n");
            return -1;
        }

        if (!bfd_get_section_contents(bfd_h, bfd_sec, sec->bytes, 0, size)) { // Load section contents
            fprintf(stderr, "failed to read section '%s' (%s)\n",
                    secname, bfd_errmsg(bfd_get_error()));
            return -1;
        }
    }

    return 0;
}

// Function to load the entire binary using BFD
static int load_binary_bfd(std::string &fname, Binary *bin, Binary::BinaryType type) {
    int ret;
    bfd *bfd_h;
    const bfd_arch_info_type *bfd_info;

    bfd_h = NULL;

    bfd_h = open_bfd(fname); // Open the binary file
    if (!bfd_h) {
        goto fail;
    }

    bin->filename = std::string(fname);
    bin->entry = bfd_get_start_address(bfd_h); // Get entry point of the binary

    bin->type_str = std::string(bfd_h->xvec->name);
    switch (bfd_h->xvec->flavour) { // Determine binary type (ELF, PE, etc.)
    case bfd_target_elf_flavour:
        bin->type = Binary::BIN_TYPE_ELF;
        break;
    case bfd_target_coff_flavour:
        bin->type = Binary::BIN_TYPE_PE;
        break;
    case bfd_target_unknown_flavour:
    default:
        fprintf(stderr, "unsupported binary type (%s)\n", bfd_h->xvec->name);
        goto fail;
    }

    bfd_info = bfd_get_arch_info(bfd_h); // Get architecture information
    bin->arch_str = std::string(bfd_info->printable_name);
    switch (bfd_info->mach) {
    case bfd_mach_i386_i386:
        bin->arch = Binary::ARCH_X86; 
        bin->bits = 32;
        break;
    case bfd_mach_x86_64:
        bin->arch = Binary::ARCH_X86;
        bin->bits = 64;
        break;
    default:
        fprintf(stderr, "unsupported architecture (%s)\n",
                bfd_info->printable_name);
        goto fail;
    }

    /* Symbol handling is best-effort only (they may not even be present) */
    load_symbols_bfd(bfd_h, bin); // Load symbols from the symbol table
    load_dynsym_bfd(bfd_h, bin); // Load dynamic symbols

    if (load_sections_bfd(bfd_h, bin) < 0) goto fail; // Load sections

    ret = 0;
    goto cleanup;

fail:
    ret = -1;

cleanup:
    if (bfd_h) bfd_close(bfd_h); // Close BFD handle

    return ret;
}

// Wrapper function to load a binary
int load_binary(std::string &fname, Binary *bin, Binary::BinaryType type) {
    return load_binary_bfd(fname, bin, type);
}

// Function to unload the binary and free allocated memory
void unload_binary(Binary *bin) {
    size_t i;
    Section *sec;

    for (i = 0; i < bin->sections.size(); i++) {
        sec = &bin->sections[i];
        if (sec->bytes) {
            free(sec->bytes); // Free memory allocated for section contents
        }
    }
}
