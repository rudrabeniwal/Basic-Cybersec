/* Linearly disassemble a given binary using Capstone. */

#include <stdio.h>
#include <string>
#include <capstone/capstone.h>
#include "loader.h"

/* Function to disassemble the binary.
 * Parameters:
 *   bin - pointer to the binary structure.
 * Returns:
 *   0 on success, -1 on failure.
 */
int disasm(Binary *bin)
{
  csh dis;                  // Handle for the Capstone disassembler.
  cs_insn *insns;           // Array to store disassembled instructions.
  Section *text;            // Pointer to the text section of the binary.
  size_t n;                 // Number of disassembled instructions.

  /* Retrieve the text section of the binary.
   * The text section contains the executable code.
   */
  text = bin->get_text_section();
  if (!text) {
    fprintf(stderr, "Nothing to disassemble\n");
    return 0;
  }

  /* Initialize the Capstone disassembler for the x86_64 architecture.
   * CS_ARCH_X86 indicates x86 architecture.
   * CS_MODE_64 indicates 64-bit mode.
   * The handle 'dis' is used for subsequent operations.
   */
  if (cs_open(CS_ARCH_X86, CS_MODE_64, &dis) != CS_ERR_OK) {
    fprintf(stderr, "Failed to open Capstone\n");
    return -1;
  }

  /* Disassemble the text section.
   * text->bytes contains the raw bytes of the text section.
   * text->size is the size of the text section.
   * text->vma is the starting virtual memory address of the text section.
   * 0 indicates disassemble all available instructions.
   * insns is populated with the disassembled instructions.
   * n is the number of instructions disassembled.
   */
  n = cs_disasm(dis, text->bytes, text->size, text->vma, 0, &insns);
  if (n <= 0) {
    fprintf(stderr, "Disassembly error: %s\n", cs_strerror(cs_errno(dis)));
    return -1;
  }

  /* Iterate over the disassembled instructions and print them.
   * The instruction address, raw bytes, mnemonic, and operands are printed.
   */
  for (size_t i = 0; i < n; i++) {
    printf("0x%016jx: ", insns[i].address);
    for (size_t j = 0; j < 16; j++) {
      if (j < insns[i].size) printf("%02x ", insns[i].bytes[j]);
      else printf("   ");
    }
    printf("%-12s %s\n", insns[i].mnemonic, insns[i].op_str);
  }

  /* Free the memory allocated for the disassembled instructions.
   * Close the Capstone handle.
   */
  cs_free(insns, n);
  cs_close(&dis);

  return 0;
}

/* Main function to load a binary file and disassemble it.
 * Parameters:
 *   argc - argument count.
 *   argv - argument vector.
 * Returns:
 *   0 on success, 1 on failure.
 */
int main(int argc, char *argv[])
{
  Binary bin;                // Binary structure to hold the loaded binary.
  std::string fname;         // File name of the binary to be loaded.

  /* Check if the binary file name is provided.
   * If not, print usage information and return an error code.
   */
  if (argc < 2) {
    printf("Usage: %s <binary>\n", argv[0]);
    return 1;
  }

  /* Assign the provided binary file name to 'fname'.
   * Load the binary into the 'bin' structure.
   * Binary::BIN_TYPE_AUTO indicates automatic detection of the binary type.
   * If loading fails, return an error code.
   */
  fname.assign(argv[1]);
  if (load_binary(fname, &bin, Binary::BIN_TYPE_AUTO) < 0) {
    return 1;
  }

  /* Disassemble the loaded binary.
   * If disassembly fails, return an error code.
   */
  if (disasm(&bin) < 0) {
    return 1;
  }

  /* Unload the binary and free associated resources.
   */
  unload_binary(&bin);

  return 0;
}

/*
compile it using - g++ -std=c++11 -o basic_capstone_linear basic_capstone_linear.cc loader.o -lbfd -lcapstone
*/