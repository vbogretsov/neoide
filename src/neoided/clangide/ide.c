#include "ide.h"

#include <stdbool.h>
#include <stdlib.h>
#include <string.h>

#include <errno.h>  //TODO: remove.
#include <stdio.h>  //TODO: remove.

#include <clang-c/Index.h>

#include "hashmap.h"
#include "libclang.h"

static const unsigned TRANSLATION_OPTIONS =
    CXTranslationUnit_PrecompiledPreamble
    | CXTranslationUnit_CacheCompletionResults
    | CXTranslationUnit_Incomplete;

static const unsigned COMPLETION_OPTIONS =
    CXCodeComplete_IncludeMacros
    | CXCodeComplete_IncludeBriefComments
    | CXCodeComplete_IncludeCodePatterns;

typedef void (*complete_chunk_t)(
    completion_t*,
    unsigned*,
    unsigned*,
    unsigned*,
    unsigned*,
    const char*);

struct ide
{
    const char* const* flags;
    unsigned nflags;
    CXIndex index;
    libclang_t* libclang;
    hashmap_t* units;
    hashmap_t* kind_chars;
    hashmap_t* kind_names;
    hashmap_t* completion_chunks;
};

struct completions
{
    CXCodeCompleteResults* handle;
    // TODO: avoid ide_t reference. Currently this reference is required
    // because we need action tables (ide_t.kind_chars, ide_t.kind_names,
    // ide_t.completion_chunks) libclang functions.
    ide_t* ide;
};

static int string_hash(const void* string)
{
    int hash = 1;

    for (char const* p = string; p != NULL && *p != '\0'; ++p)
    {
        hash = (hash << 1) ^ *p;
    }

    return hash;
}

static bool string_equals(const void* a, const void* b)
{
    return strcmp(a, b) == 0;
}

static int unsigned_hash(const void* value)
{
    return (unsigned)value;
}

static bool unsigned_equals(const void* a, const void* b)
{
    return (unsigned)a == (unsigned)b;
}

static void dispose_unit(void* ctx, const void* filename, void* unit)
{
    free((void*)filename);
    ((ide_t*)ctx)->libclang->dispose_tu((CXTranslationUnit)unit);
}

static hashmap_t* init_kind_chars()
{
    hashmap_t* map = hashmap_alloc(&unsigned_hash, &unsigned_equals);

    hashmap_set(map, (void*)CXCursor_StructDecl, (void*)'t');
    hashmap_set(map, (void*)CXCursor_UnionDecl, (void*)'t');
    hashmap_set(map, (void*)CXCursor_EnumConstantDecl, (void*)'t');
    hashmap_set(map, (void*)CXCursor_EnumDecl, (void*)'t');
    hashmap_set(map, (void*)CXCursor_TypedefDecl, (void*)'t');
    hashmap_set(map, (void*)CXCursor_ClassTemplate, (void*)'t');
    hashmap_set(map, (void*)CXCursor_ClassDecl, (void*)'t');
    hashmap_set(map, (void*)CXCursor_ConversionFunction, (void*)'f');
    hashmap_set(map, (void*)CXCursor_FunctionTemplate, (void*)'f');
    hashmap_set(map, (void*)CXCursor_FunctionDecl, (void*)'f');
    hashmap_set(map, (void*)CXCursor_Constructor, (void*)'m');
    hashmap_set(map, (void*)CXCursor_Destructor, (void*)'m');
    hashmap_set(map, (void*)CXCursor_CXXMethod, (void*)'m');
    hashmap_set(map, (void*)CXCursor_FieldDecl, (void*)'m');
    hashmap_set(map, (void*)CXCursor_VarDecl, (void*)'v');
    hashmap_set(map, (void*)CXCursor_TemplateTypeParameter, (void*)'p');
    hashmap_set(map, (void*)CXCursor_ParmDecl, (void*)'s');
    hashmap_set(map, (void*)CXCursor_PreprocessingDirective, (void*)'D');
    hashmap_set(map, (void*)CXCursor_MacroDefinition, (void*)'M');

    return map;
}

static hashmap_t* init_kind_names()
{
    hashmap_t* map = hashmap_alloc(&unsigned_hash, &unsigned_equals);

    hashmap_set(map, (void*)CXCursor_StructDecl, (void*)"struct ");
    hashmap_set(map, (void*)CXCursor_UnionDecl, (void*)"union ");
    hashmap_set(map, (void*)CXCursor_EnumConstantDecl, (void*)"enum ");
    hashmap_set(map, (void*)CXCursor_EnumDecl, (void*)"enum ");
    hashmap_set(map, (void*)CXCursor_TypedefDecl, (void*)"typedef ");
    hashmap_set(map, (void*)CXCursor_ClassTemplate, (void*)"class ");
    hashmap_set(map, (void*)CXCursor_ClassDecl, (void*)"class ");

    return map;
}

static void buffcpy(
    char buff[],
    unsigned* buff_pos,
    unsigned buff_size,
    const char* str)
{
    unsigned i = 0;
    for (;i < buff_size && str[i] != '\0'; ++i)
    {
        buff[*buff_pos + i] = str[i];
    }

    buff[*buff_pos + i] = '\0';
    *buff_pos += i;
}

static void complete_typed_text(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
    unsigned* menu_i,
    unsigned* sort_i,
    const char* part)
{
    buffcpy(completion->word, word_i, WORD_SIZE, part);
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, part);
    buffcpy(completion->sort, sort_i, SORT_SIZE, part);
}

static void complete_text(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
    unsigned* menu_i,
    unsigned* sort_i,
    const char* part)
{
    buffcpy(completion->word, word_i, WORD_SIZE, part);
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, part);
    buffcpy(completion->sort, sort_i, SORT_SIZE, part);
}

static void complete_placeholder(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
    unsigned* menu_i,
    unsigned* sort_i,
    const char* part)
{
    // buffcpy(completion->word, word_i, WORD_SIZE, "<#");
    // buffcpy(completion->word, word_i, WORD_SIZE, part);
    // buffcpy(completion->word, word_i, WORD_SIZE, "#>");
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, part);
}

static void complete_result_type(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
    unsigned* menu_i,
    unsigned* sort_i,
    const char* part)
{
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, part);
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, " ");
}

static void complete_symbol(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
    unsigned* menu_i,
    unsigned* sort_i,
    const char* part)
{
    // buffcpy(completion->word, word_i, WORD_SIZE, part);
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, part);
}

static hashmap_t* init_completion_chunks()
{
    hashmap_t* map = hashmap_alloc(&unsigned_hash, &unsigned_equals);

    hashmap_set(map, (void*)CXCompletionChunk_TypedText, &complete_typed_text);
    hashmap_set(map, (void*)CXCompletionChunk_Text, &complete_text);
    hashmap_set(map, (void*)CXCompletionChunk_ResultType, &complete_result_type);
    hashmap_set(map, (void*)CXCompletionChunk_Placeholder, &complete_placeholder);
    hashmap_set(map, (void*)CXCompletionChunk_LeftParen, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_RightParen, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_LeftBracket, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_RightBracket, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_LeftBrace, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_RightBrace, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_LeftAngle, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_RightAngle, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_Comma, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_Colon, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_SemiColon, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_Equal, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_HorizontalSpace, &complete_symbol);
    hashmap_set(map, (void*)CXCompletionChunk_VerticalSpace, &complete_symbol);

    return map;
}

ide_t* ide_alloc(const char* libclang_path)
{
    libclang_t* libclang = libclang_load(libclang_path);

    if (libclang == NULL)
    {
        return NULL;
    }

    ide_t* ide = (ide_t*)malloc(sizeof(ide_t));

    ide->libclang = libclang;
    ide->flags = NULL;
    ide->nflags = 0;
    ide->index = libclang->create_index(1, 0);
    ide->units = hashmap_alloc(&string_hash, &string_equals);
    ide->kind_chars = init_kind_chars();
    ide->kind_names = init_kind_names();
    ide->completion_chunks = init_completion_chunks();

    return ide;
}

void ide_free(ide_t* ide)
{
    hashmap_free(ide->completion_chunks);
    hashmap_free(ide->kind_names);
    hashmap_free(ide->kind_chars);
    hashmap_each(ide->units, ide, &dispose_unit);
    hashmap_free(ide->units);
    ide->libclang->dispose_index(ide->index);
    libclang_close(ide->libclang);
    free(ide);
}

void ide_on_file_open(ide_t* ide, const char* filename)
{
    void* unit;
    if (!hashmap_get(ide->units, filename, &unit))
    {
        // TODO: handle errors
        unit = ide->libclang->parse_tu(
            ide->index,
            filename,
            ide->flags,
            ide->nflags,
            NULL,
            0,
            TRANSLATION_OPTIONS);
        const char* key = strdup(filename);
        hashmap_set(ide->units, key, unit);
    }
}

void ide_on_file_close(ide_t* ide, const char* filename)
{
    void* unit;
    if (!hashmap_get(ide->units, filename, &unit))
    {
        ide->libclang->dispose_tu((CXTranslationUnit)unit);
    }
}

void ide_on_file_save(ide_t* ide, const char* filename)
{
    void* unit;
    if (!hashmap_get(ide->units, filename, &unit))
    {
        ide->libclang->reparse_tu(
            (CXTranslationUnit)unit,
            0,
            NULL,
            TRANSLATION_OPTIONS);
    }
}

static void read_completion(
    ide_t* ide,
    CXCompletionResult* result,
    unsigned index,
    void* ctx,
    void (*oncompletion)(unsigned, void*, completion_t*))
{
    completion_t completion;
    completion.abbr[0] = '\0';
    completion.word[0] = '\0';
    completion.menu[0] = '\0';
    completion.sort[0] = '\0';
    unsigned abbr_i = 0;
    unsigned word_i = 0;
    unsigned menu_i = 0;
    unsigned sort_i = 0;

    void* kind_char;
    if (hashmap_get(ide->kind_chars, (void*)result->CursorKind, &kind_char))
    {
        completion.kind = (char)kind_char;
    }

    void* kind_name;
    if (hashmap_get(ide->kind_names, (void*)result->CursorKind, &kind_name))
    {
        buffcpy(completion.abbr, &abbr_i, ABBR_SIZE, (char*)kind_name);
    }

    CXCompletionString comp_string = result->CompletionString;
    unsigned num_chunks =
        ide->libclang->get_num_completion_chunks(comp_string);

    for (unsigned i = 0; i < num_chunks; ++i)
    {
        CXString chunk_text =
            ide->libclang->get_completion_chunk_text(comp_string, i);

        enum CXCompletionChunkKind chunk_kind =
            ide->libclang->get_completion_chunk_kind(comp_string, i);

        const char* part = ide->libclang->get_string(chunk_text);

        void* chunk_fn;
        if (hashmap_get(ide->completion_chunks, (void*)chunk_kind, &chunk_fn))
        {
            (*(complete_chunk_t)chunk_fn)(
                &completion, &abbr_i, &word_i, &menu_i, &sort_i, part);
        }

        ide->libclang->dispose_string(chunk_text);
    }

    (*oncompletion)(index, ctx, &completion);
}

// NOTE: this method should be general and operate native clang API types, but
// we're doing this completer only for VIM and for simplicity and performance
// reasons we're translating clang completions to VIM complete-item here.
completions_t* ide_find_completions(
    ide_t* ide,
    const char* filename,
    unsigned line,
    unsigned column,
    const char* content,
    unsigned size)
{
    void* unit;
    if (!hashmap_get(ide->units, filename, &unit))
    {
        // TODO: add error details.
        return NULL;
    }

    struct CXUnsavedFile unsaved_file =
        {.Filename = filename, .Contents = content, .Length = size};

    CXTranslationUnit tu = (CXTranslationUnit)unit;
    // TODO: dispose completions
    CXCodeCompleteResults* results = ide->libclang->complete_at(
        tu,
        filename,
        line,
        column,
        (struct CXUnsavedFile[]){unsaved_file},
        1,
        COMPLETION_OPTIONS);

    if (!results)
    {
        // TODO: add error details.
        return NULL;
    }

    completions_t* completions = (completions_t*)malloc(sizeof(completions_t));
    completions->handle = results;
    completions->ide = ide;

    return completions;
}


unsigned completions_count(completions_t* completions)
{
    return completions->handle->NumResults;
}

void completions_iter(
    completions_t* completions,
    void* ctx,
    void (*action)(unsigned, void*, completion_t*))
{
    for (unsigned i = 0; i < completions->handle->NumResults; ++i)
    {
        read_completion(
            completions->ide,
            &(completions->handle->Results[i]),
            i,
            ctx,
            action);
    }
}

void completions_free(completions_t* completions)
{
    completions->ide->libclang->dispose_completion(completions->handle);
    free(completions);
}