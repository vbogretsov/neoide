#include "libclang.h"

#include <errno.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "import.h"

char ERROR[1024];

#define IMPORT_FUNCTION(libclang, member, type, name)                         \
    if (!(libclang->member = (type)import_name(libclang->handle, name)))      \
    {                                                                         \
        sprintf(ERROR, "Unable to import function %s\n", name);               \
        libclang_free(libclang);                                              \
        return NULL;                                                          \
    }

// https://clang.llvm.org/doxygen/group__CINDEX.html#ga51eb9b38c18743bf2d824c6230e61f93
typedef CXIndex (*clang_create_index_t)(int, int);

// https://clang.llvm.org/doxygen/group__CINDEX.html#ga166ab73b14be73cbdcae14d62dbab22a
typedef void (*clang_dispose_index_t)(CXIndex);

// https://clang.llvm.org/doxygen/group__CINDEX__TRANSLATION__UNIT.html#ga2baf83f8c3299788234c8bce55e4472e
typedef CXTranslationUnit (*clang_parse_tu_t)(
    CXIndex, const char*, const char* const*, unsigned, struct CXUnsavedFile*,
    unsigned, unsigned);

// https://clang.llvm.org/doxygen/group__CINDEX__TRANSLATION__UNIT.html#ga524e76bf2a809d037934d4be51ea448a
typedef int (*clang_reparse_tu_t)(
    CXTranslationUnit, unsigned, struct CXUnsavedFile*, unsigned);

// https://clang.llvm.org/doxygen/group__CINDEX__TRANSLATION__UNIT.html#gaee753cb0036ca4ab59e48e3dff5f530a
typedef void (*clang_dispose_tu_t)(CXTranslationUnit);

// https://clang.llvm.org/doxygen/group__CINDEX__CODE__COMPLET.html#ga50fedfa85d8d1517363952f2e10aa3bf
typedef CXCodeCompleteResults* (*clang_complete_at_t)(
    CXTranslationUnit, const char*, unsigned, unsigned, struct CXUnsavedFile*,
    unsigned, unsigned);

// https://clang.llvm.org/doxygen/group__CINDEX__CODE__COMPLET.html#ga206cc6ea7be311537bb0fab584ebc6c1
typedef void (*clang_dsipose_completion_t)(CXCodeCompleteResults*);

// https://clang.llvm.org/doxygen/group__CINDEX__STRING.html#gafd043aa189e990b9e327e9f95a1da8a5
typedef const char* (*clang_get_string_t)(CXString);

// https://clang.llvm.org/doxygen/group__CINDEX__STRING.html#gaeff715b329ded18188959fab3066048f
typedef void (*clang_dispose_string_t)(CXString);

// https://clang.llvm.org/doxygen/group__CINDEX__CODE__COMPLET.html#ga46e843acdf63d9a7a0c7341a2d222c49
typedef unsigned (*clang_get_completion_priority_t)(CXCompletionString);

// https://clang.llvm.org/doxygen/group__CINDEX__CODE__COMPLET.html#ga32163145c7f0013e5f2ac7176a8ee0ed
typedef CXString (*clang_get_completion_brief_comment_t)(CXCompletionString);

// https://clang.llvm.org/doxygen/group__CINDEX__CODE__COMPLET.html#ga76018aa1a7225268546e4d75dca5dbce
typedef unsigned (*clang_get_num_completion_chunks_t)(CXCompletionString);

// https://clang.llvm.org/doxygen/group__CINDEX__CODE__COMPLET.html#ga98d4c869dda8fd4b5386f62d02d6ba0b
typedef CXString (*clang_get_completion_chunk_text_t)(
    CXCompletionString, unsigned);

// https://clang.llvm.org/doxygen/group__CINDEX__CODE__COMPLET.html#gac61e18c6d895d85f1476c6091d486091
typedef enum CXCompletionChunkKind (*clang_get_completion_chunk_kind_t)(
    CXCompletionString,
    unsigned);

// https://clang.llvm.org/doxygen/group__CINDEX__CODE__COMPLET.html#gadb669685b9ef1f8ca62b2a044b846ac1
typedef unsigned (*clang_default_code_complete_options_t)();

struct libclang
{
    void* handle;
    clang_create_index_t create_index;
    clang_dispose_index_t dispose_index;
    clang_parse_tu_t parse_tu;
    clang_reparse_tu_t reparse_tu;
    clang_dispose_tu_t dispose_tu;
    clang_complete_at_t complete_at;
    clang_dsipose_completion_t dispose_completion;
    clang_get_string_t get_string;
    clang_dispose_string_t dispose_string;
    clang_get_completion_priority_t get_completion_priority;
    clang_get_completion_brief_comment_t get_completion_brief_comment;
    clang_get_num_completion_chunks_t get_num_completion_chunks;
    clang_get_completion_chunk_text_t get_completion_chunk_text;
    clang_get_completion_chunk_kind_t get_completion_chunk_kind;
    clang_default_code_complete_options_t default_code_complete_options;
};

typedef void (*complete_chunk_t)(
    completion_t*, unsigned*, unsigned*, const char*);

char** make_string_array(unsigned size) {
    return malloc(sizeof(char*) * size);
}

void free_string_array(char** array, unsigned size) {
    for (unsigned i = 0; i < size; ++i) {
        free(array[i]);
    }
    free(array);
}

const char* libclang_error()
{
    return errno < 0 ? ERROR : strerror(errno);
}

libclang_t* libclang_load(const char* path)
{
    void* handle = load_library(path);

    if (!handle)
    {
        errno = ENOENT;
        return NULL;
    }

    libclang_t* so = (libclang_t*)malloc(sizeof(libclang_t));
    so->handle = handle;

    IMPORT_FUNCTION(so, create_index, clang_create_index_t,
                    "clang_createIndex");
    IMPORT_FUNCTION(so, dispose_index, clang_dispose_index_t,
                    "clang_disposeIndex");
    IMPORT_FUNCTION(so, parse_tu, clang_parse_tu_t,
                    "clang_parseTranslationUnit");
    IMPORT_FUNCTION(so, reparse_tu, clang_reparse_tu_t,
                    "clang_reparseTranslationUnit");
    IMPORT_FUNCTION(so, dispose_tu, clang_dispose_tu_t,
                    "clang_disposeTranslationUnit");
    IMPORT_FUNCTION(so, complete_at, clang_complete_at_t,
                    "clang_codeCompleteAt");
    IMPORT_FUNCTION(so, dispose_completion, clang_dsipose_completion_t,
                    "clang_disposeCodeCompleteResults");
    IMPORT_FUNCTION(so, get_string, clang_get_string_t,
                    "clang_getCString");
    IMPORT_FUNCTION(so, dispose_string, clang_dispose_string_t,
                    "clang_disposeString");
    IMPORT_FUNCTION(so, get_completion_priority,
                    clang_get_completion_priority_t,
                    "clang_getCompletionPriority");
    IMPORT_FUNCTION(so, get_num_completion_chunks,
                    clang_get_num_completion_chunks_t,
                    "clang_getNumCompletionChunks");
    IMPORT_FUNCTION(so, get_completion_chunk_text,
                    clang_get_completion_chunk_text_t,
                    "clang_getCompletionChunkText");
    IMPORT_FUNCTION(so, get_completion_chunk_kind,
                    clang_get_completion_chunk_kind_t,
                    "clang_getCompletionChunkKind");

    return so;
}

void libclang_free(libclang_t* so)
{
    close_library(so->handle);
    free(so);
}

index_t libclang_create_index(
    libclang_t* so, int excludeDeclarationsFromPCH, int displayDiagnostics)
{
    return so->create_index(excludeDeclarationsFromPCH, displayDiagnostics);
}

void libclang_dispose_index(libclang_t* so, index_t index)
{
    so->dispose_index(index);
}

translation_unit_t libclang_parse_tu(
    libclang_t* so, index_t index, const char* path, const char* const* flags,
    unsigned num_flags, unsigned options)
{
    return so->parse_tu(index, path, flags, num_flags, NULL, 0, options);
}

int libclang_reparse_tu(
    libclang_t* so, translation_unit_t tu, unsigned options)
{
    return so->reparse_tu(tu, 0, NULL, options);
}

void libclang_dispose_tu(libclang_t* so, translation_unit_t tu)
{
    so->dispose_tu(tu);
}

static void buffcpy(char buff[], unsigned* pos, unsigned size, const char* str)
{
    unsigned i = 0;
    for (;i < size && str[i] != '\0'; ++i)
    {
        buff[*pos + i] = str[i];
    }

    buff[*pos + i] = '\0';
    *pos += i;
}

static void complete_typed_text(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
    const char* part)
{
    buffcpy(completion->word, word_i, WORD_SIZE, part);
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, part);
}

static void complete_text(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
    const char* part)
{
    buffcpy(completion->word, word_i, WORD_SIZE, part);
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, part);
}

static void complete_placeholder(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
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
    const char* part)
{
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, part);
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, " ");
}

static void complete_symbol(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
    const char* part)
{
    // buffcpy(completion->word, word_i, WORD_SIZE, part);
    buffcpy(completion->abbr, abrr_i, ABBR_SIZE, part);
}

static void skipp_completion(
    completion_t* completion,
    unsigned* abrr_i,
    unsigned* word_i,
    const char* part)
{
}

static char kind_char(enum CXCursorKind kind)
{
    switch (kind)
    {
        case CXCursor_StructDecl:
            return 't';
        case CXCursor_UnionDecl:
            return 't';
        case CXCursor_EnumConstantDecl:
            return 't';
        case CXCursor_EnumDecl:
            return 't';
        case CXCursor_TypedefDecl:
            return 't';
        case CXCursor_ClassTemplate:
            return 't';
        case CXCursor_ClassDecl:
            return 't';
        case CXCursor_ConversionFunction:
            return 'f';
        case CXCursor_FunctionTemplate:
            return 'f';
        case CXCursor_FunctionDecl:
            return 'f';
        case CXCursor_Constructor:
            return 'm';
        case CXCursor_Destructor:
            return 'm';
        case CXCursor_CXXMethod:
            return 'm';
        case CXCursor_FieldDecl:
            return 'm';
        case CXCursor_VarDecl:
            return 'v';
        case CXCursor_TemplateTypeParameter:
            return 'p';
        case CXCursor_ParmDecl:
            return 's';
        case CXCursor_PreprocessingDirective:
            return 'D';
        case CXCursor_MacroDefinition:
            return 'M';
        default:
            return '?';
    }
}

static const char* kind_name(enum CXCursorKind kind)
{
    switch (kind)
    {
        case CXCursor_StructDecl:
            return "struct ";
        case CXCursor_UnionDecl:
            return "union ";
        case CXCursor_EnumConstantDecl:
            return "enum ";
        case CXCursor_EnumDecl:
            return "enum ";
        case CXCursor_TypedefDecl:
            return "typedef ";
        case CXCursor_ClassTemplate:
            return "class ";
        case CXCursor_ClassDecl:
            return "class ";
        default:
            return "";
    }
}

static complete_chunk_t completer(enum CXCompletionChunkKind kind)
{
    switch (kind)
    {
        case CXCompletionChunk_TypedText:
            return &complete_typed_text;
        case CXCompletionChunk_Text:
            return &complete_text;
        case CXCompletionChunk_ResultType:
            return &complete_result_type;
        case CXCompletionChunk_Placeholder:
            return &complete_placeholder;
        case CXCompletionChunk_LeftParen:
            return &complete_symbol;
        case CXCompletionChunk_RightParen:
            return &complete_symbol;
        case CXCompletionChunk_LeftBracket:
            return &complete_symbol;
        case CXCompletionChunk_RightBracket:
            return &complete_symbol;
        case CXCompletionChunk_LeftBrace:
            return &complete_symbol;
        case CXCompletionChunk_RightBrace:
            return &complete_symbol;
        case CXCompletionChunk_LeftAngle:
            return &complete_symbol;
        case CXCompletionChunk_RightAngle:
            return &complete_symbol;
        case CXCompletionChunk_Comma:
            return &complete_symbol;
        case CXCompletionChunk_Colon:
            return &complete_symbol;
        case CXCompletionChunk_SemiColon:
            return &complete_symbol;
        case CXCompletionChunk_Equal:
            return &complete_symbol;
        case CXCompletionChunk_HorizontalSpace:
            return &complete_symbol;
        case CXCompletionChunk_VerticalSpace:
            return &complete_symbol;
        default:
            return &skipp_completion;
    }
}

static void visit_completion(
    libclang_t* so, CXCompletionResult* result, unsigned i, void* ctx,
    void(*action)(completion_t*, unsigned, void*))
{
    completion_t completion;
    completion.abbr[0] = '\0';
    completion.word[0] = '\0';
    unsigned abbr_i = 0;
    unsigned word_i = 0;

    completion.kind = kind_char(result->CursorKind);
    buffcpy(
        completion.abbr, &abbr_i, ABBR_SIZE, kind_name(result->CursorKind));
    CXCompletionString comp_string = result->CompletionString;
    unsigned num_chunks = so->get_num_completion_chunks(comp_string);

    for (unsigned j = 0; j < num_chunks; ++j)
    {
        CXString chunk_text = so->get_completion_chunk_text(
            comp_string, j);
        enum CXCompletionChunkKind kind = so->get_completion_chunk_kind(
            comp_string, j);
        const char* part = so->get_string(chunk_text);
        (*completer(kind))(&completion, &abbr_i, &word_i, part);
        so->dispose_string(chunk_text);
    }

    (*action)(&completion, i, ctx);
}

completion_results_t* libclang_complete_at(
    libclang_t* so, translation_unit_t tu, unsigned options,
    const char* file_path, const char* file_content, unsigned file_size,
    unsigned line, unsigned column)
{
    struct CXUnsavedFile unsaved_file =
        {.Filename = file_path, .Contents = file_content, .Length = file_size};

    return so->complete_at(
        tu, file_path, line, column,
        (struct CXUnsavedFile[]){unsaved_file}, 1, options);
}

void libclang_completions_free(libclang_t* so, completion_results_t* results)
{
    so->dispose_completion(results);
}

void libclang_completions_foreach(
    libclang_t* so, completion_results_t* results, void* ctx,
    void(*action)(completion_t*, unsigned, void*))
{
    for (unsigned i = 0; i < results->NumResults; ++i)
    {
        visit_completion(so, &(results->Results[i]), i, ctx, action);
    }
}
