/**
 * Interface for libclang loaded dynamically.
 *
 * Jun 29 2017 Vladimir Bogretsov <bogrecov@gmail.com>
 */
#ifndef LIBCLANG_H
#define LIBCLANG_H

#include "clang-c/Index.h"

typedef struct libclang libclang_t;
typedef CXIndex index_t;
typedef CXTranslationUnit translation_unit_t;
typedef CXCodeCompleteResults completion_results_t;

#define ABBR_SIZE 128
#define WORD_SIZE 128

typedef struct
{
    char abbr[ABBR_SIZE];
    char word[WORD_SIZE];
    char kind;
    unsigned priority;
} completion_t;

/**
 * Create uninitialized char array.
 * @param  size Array size.
 * @return      Pointer to array allocated.
 */
char** make_string_array(unsigned size);

/**
 * Free strings array.
 * @param array Array to be deallocated.
 * @param size  Array size.
 */
void free_string_array(char** array, unsigned size);

/**
 * Get the error message.
 * @param  errornum Error number.
 * @return          Error message.
 */
const char* libclang_error();

/**
 * Load libclang shared library.
 * @param  path Shared library path.
 * @return      Library handle.
 */
libclang_t* libclang_load(const char* path);

/**
 * Unload libclang shared lbrary.
 * @param so Library handle.
 */
void libclang_free(libclang_t* so);

/**
 * Create clang index.
 * @param  so                         Library handle.
 * @param  excludeDeclarationsFromPCH See clang C API clang_createIndex.
 * @param  displayDiagnostics         See clang C API clang_createIndex.
 * @return                            Index created.
 */
index_t libclang_create_index(
    libclang_t* so, int excludeDeclarationsFromPCH, int displayDiagnostics);

/**
 * Dispose clang index.
 * @param so    Lbrary handle.
 * @param index Index to dispose.
 */
void libclang_dispose_index(libclang_t* so, index_t index);

/**
 * Parse translation unit.
 * @param  so        Library handle.
 * @param  index     Clang index.
 * @param  path      Source file path.
 * @param  flags     Compiler flags.
 * @param  num_flags Number of compiler flags.
 * @param  options   Translation options.
 * @return           Translation unit parsed.
 */
translation_unit_t libclang_parse_tu(
    libclang_t* so, index_t index, const char* path,
    const char* const* flags, unsigned num_flags, unsigned options);

/**
 * Reparse translation unit.
 * @param  so      Library handle.
 * @param  tu      Translation unit to reparse.
 * @param  options Parse options.
 * @return         0 if success.
 */
int libclang_reparse_tu(
    libclang_t* so, translation_unit_t tu, unsigned options);

/**
 * Dispose translation unit.
 * @param so Library handle.
 * @param tu Translation unit to dispose.
 */
void libclang_dispose_tu(libclang_t* so, translation_unit_t tu);

/**
 * Get autocompletions in the file provided.
 * @param  so           Library handle.
 * @param  tu           Translation unit.
 * @param  options      Completion options.
 * @param  file_path    Path of file for which to search completions.
 * @param  file_content Content of file for which to search completions.
 * @param  file_size    Size of file for which to search completions.
 * @param  line         Line number where to search completions.
 * @param  column       Column number where to search completions.
 * @return              Completions results.
 */
completion_results_t* libclang_complete_at(
    libclang_t* so, translation_unit_t tu, unsigned options,
    const char* file_path, const char* file_content, unsigned file_size,
    unsigned line, unsigned column);

/**
 * Free completion results.
 * @param  so     Library handle.
 * @param results Results to be deallocated.
 */
void libclang_completions_free(libclang_t* so, completion_results_t* results);

/**
 * Iterate over completion results.
 * @param  so     Library handle.
 * @param results Completion results.
 * @param ctx     Closure context.
 * @param action  Acition to apply for each completion result.
 */
void libclang_completions_foreach(
    libclang_t* so, completion_results_t* results, void* ctx,
    void(*action)(completion_t*, unsigned, void*));

#endif // !LIBCLANG_H