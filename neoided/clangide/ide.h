/**
 * Interface for IDE C/C++ plugin based on libclang.
 *
 * Jun 7 2017 Vladimir Bogretsov <bogrecov@gmail.com>
 */
#ifndef IDE_H
#define IDE_H

typedef struct ide ide_t;

typedef struct completions completions_t;

#define ABBR_SIZE 128
#define MENU_SIZE 128  // TODO: remove menu member.
#define SORT_SIZE 128
#define WORD_SIZE 128
#define INFO_SIZE 1024

typedef struct
{
    char abbr[ABBR_SIZE];
    char menu[MENU_SIZE];  // TODO: remove menu member.
    char sort[SORT_SIZE];
    char word[WORD_SIZE];
    char info[INFO_SIZE];
    char kind;
    unsigned priority;
} completion_t;

typedef struct
{
    const char* filename;
    unsigned line;
    unsigned column;
} location_t;

/**
 * Allocate and initialize new ide instance.
 * @param  libclang_path Path to libclang library.
 * @param  flags         Compiler flags.
 * @param  nflags        Number of compiler flags.
 * @return               Initialized ide instance.
 */
ide_t* ide_alloc(const char* libclang_path);

/**
 * Deallocate the ide instance provided.
 * @param ide Instance to be deallocated.
 */
void ide_free(ide_t* ide);

/**
 * Notify IDE about opening a file.
 * @param ide      IDE instance.
 * @param filename Opened file name.
 */
void ide_on_file_open(ide_t* ide, const char* filename);

/**
 * Notifgy IDE about file save.
 * @param ide      IDE instance.
 * @param filename Saved file name.
 */
void ide_on_file_save(ide_t* ide, const char* filename);

/**
 * Notify IDE about closing a file.
 * @param ide      IDE instance.
 * @param filename Closed file name.
 */
void ide_on_file_close(ide_t* ide, const char* filename);

/**
 * Find completions for the position in the file.
 * @param ide         IDE instance.
 * @param filename    File where completions deisred.
 * @param line        Line number where completions desired.
 * @param column      Column number where completions desired.
 * @param content     Content of the file.
 * @param size        Content size.
 * @return            Cmopletion reults. The caller is responsible to delete
 *                    them.
 */
completions_t* ide_find_completions(
    ide_t* ide,
    const char* filename,
    unsigned line,
    unsigned column,
    const char* content,
    unsigned size);

/**
 * Find symbol declaration.
 * @param ide          IDE instance.
 * @param filename     File where symbol desired is located.
 * @param line         Line number where symbol desired is located.
 * @param column       Column number where symbol desired is located.
 * @param ctx          Enclosure context.
 * @param ondefinition Definition handler.
 */
void ide_find_definition(
    ide_t* ide,
    const char* filename,
    unsigned line,
    unsigned column,
    void* ctx,
    void (*ondefinition)(void*, location_t*));

/**
 * Find symbol declaration.
 * @param ide          IDE instance.
 * @param filename     File where symbol desired is located.
 * @param line         Line number where symbol desired is located.
 * @param column       Column number where symbol desired is located.
 * @param ctx          Enclosure context.
 * @param ondefinition Single declaration handler.
 */
void ide_find_declaration(
    ide_t* ide,
    const char* filename,
    unsigned line,
    unsigned column,
    void* ctx,
    void (*ondeclaration)(void*, location_t*));

/**
 * Find symbol assingment.
 * @param ide          IDE instance.
 * @param filename     File where symbol desired is located.
 * @param line         Line number where symbol desired is located.
 * @param column       Column number where symbol desired is located.
 * @param ctx          Enclosure context.
 * @param onassingment Single assingment handler.
 */
void ide_find_assingments(
    ide_t* ide,
    const char* filename,
    unsigned line,
    unsigned column,
    void* ctx,
    void (*onassingment)(void*, location_t*));

/**
 * Find symbol reference.
 * @param ide          IDE instance.
 * @param filename     File where symbol desired is located.
 * @param line         Line number where symbol desired is located.
 * @param column       Column number where symbol desired is located.
 * @param ctx          Enclosure context.
 * @param onreference  Single reference handler.
 */
void ide_find_references(
    ide_t* ide,
    const char* filename,
    unsigned line,
    unsigned column,
    void* ctx,
    void (*onreference)(void*, location_t*));

/**
 * Get completion results count.
 * @param  completions Completion resutls.
 * @return             Count of items in results.
 */
unsigned completions_count(completions_t* completions);

/**
 * Iterate over completion results and incoke the action provider for an each
 * item.
 * @param completions Completion results.
 * @param ctx         Closure context.
 * @param action      Action to apply to each completion.
 */
void completions_iter(
    completions_t* completions,
    void* ctx,
    void (*action)(unsigned, void*, completion_t*));

/**
 * Free completion results.
 * @param completions Completion results.
 */
void completions_free(completions_t* completions);

#endif // !IDE_H