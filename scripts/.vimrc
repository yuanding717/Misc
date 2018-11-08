" set VIM mode 
set nocompatible

" show existing tab with 4 spaces width
set tabstop=4
" when indenting with '>', use 4 spaces width
set shiftwidth=4
" On pressing tab, insert 4 spaces
set expandtab

" set indentation
" set autoindent
" set cindent
" set smartindent

" auto add comment on newline
set fo+=r

" enable fancy code coloring
syntax on

" set incremental (as U type) search
set incsearch

"enable modelines
set modeline

" break too long lines, visual display only
set wrap
" break at word boundary instead of anywhere in the string
set linebreak

" display the current mode and partially-typed commands in the status line
set showmode
set showcmd

" have fifty lines of command-line (etc) history
set history=100

" have the mouse enabled all the time
set mouse=a

" enable filetype detection
filetype on

" set maximum text width
set textwidth=140

" enable backspace
set backspace=indent,eol,start

" nice line numbers
set number

" show ruler
set ruler

" highlight search
set hlsearch

" enable spell check
" set spell

" utf8 encoding
set encoding=utf-8

" always show status line
set laststatus=2

" don't expand tabs for make files
autocmd FileType make set noexpandtab
