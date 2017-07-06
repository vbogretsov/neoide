if exists('g:neoide_loaded')
    finish
endif

let g:neoide_loaded = 1
let s:neoided_path =  expand('<sfile>:p:h:h') . '/bin/neoided'

function! s:start_neoide(host) abort
    return jobstart([s:neoided_path], {'rpc': v:true})
endfunction

function! s:init_go()
    call remote#host#Register('neoided', 'x', function('s:start_neoide'))
    call remote#host#RegisterPlugin('neoided', '0', [
        \ {'type': 'function', 'name': '_neoide_bufenter', 'sync': 0, 'opts': {}},
        \ {'type': 'function', 'name': '_neoide_bufsave', 'sync': 0, 'opts': {}},
        \ {'type': 'function', 'name': '_neoide_bufclose', 'sync': 0, 'opts': {}},
        \ {'type': 'function', 'name': '_neoide_find_completions', 'sync': 0, 'opts': {}},
        \ {'type': 'function', 'name': '_neoide_show_completions', 'sync': 1, 'opts': {}},
        \ {'type': 'function', 'name': '_neoide_get_completions', 'sync': 1, 'opts': {}},
        \ {'type': 'function', 'name': '_neoide_find_defenition', 'sync': 1, 'opts': {}},
        \ {'type': 'function', 'name': '_neoide_find_declaration', 'sync': 1, 'opts': {}},
        \ {'type': 'function', 'name': '_neoide_find_references', 'sync': 1, 'opts': {}},
        \ {'type': 'function', 'name': '_neoide_find_assingments', 'sync': 1, 'opts': {}}
        \ ])
endfunction

function! neoide#find_completsion() abort
    call _neoide_find_completions(&filetype, getline("."))
endfunction

function! neoide#completefunc(findstart, base) abort
    if a:findstart
        return b:complete_column
    else
        let l:completions = _neoide_get_completions(a:base)
        return {'words': l:completions, 'refresh': 'always'}
    endif
endfunction

function! neoide#force_popup() abort
    if !pumvisible()
        call _neoide_show_completions(&filetype, col('.') + 1)
    endif
endfunction

function! neoide#show_popup(position)
    if !pumvisible()
        let b:complete_column = a:position
        call feedkeys("\<C-X>\<C-U>\<C-P>", 'i')
    endif
endfunction

function! neoide#cancel_popup()
    if pumvisible()
        return 1
    else
        return 0
    endif
endfunction

function! neoide#error(message)
    echo "neoide [error]: " . a:message
endfunction

function! neoide#info(message)
    echo "neoide [info]: " . a:message
endfunction

function! s:init()
    call s:init_go()

    set completefunc=neoide#completefunc
    set completeopt+=menuone

    augroup neoide
        autocmd!
        autocmd BufEnter <buffer> call _neoide_bufenter(&filetype, expand('%:p'))
        autocmd BufLeave <buffer> call _neoide_bufclose(&filetype, expand('%:p'))
        autocmd TextChangedI <buffer> call neoide#find_completsion()
        autocmd CompleteDone <buffer> call neoide#cancel_popup()
    augroup END

    inoremap <C-Space> <C-O>:call neoide#force_popup()<CR>
    inoremap <silent> <expr> <ESC> (neoide#cancel_popup() ? "<C-E>" : "<ESC>")

    " TODO: remove
    profile start profile.log
    profile func *
    profile file *
endfunction

call s:init()

" vim:ts=4:sw=4:et