(require 'font-lock)

(defun fgscript-indent-line ()
  "Indent current line as FGScript code."
  (interactive)
  (let ((indent-level (fgscript-calculate-indentation)))
    (indent-line-to (or indent-level 0))))

(defun fgscript-calculate-indentation ()
  "Calculate the appropriate indentation level."
  (save-excursion
    (beginning-of-line)
    (cond
     ((looking-at "^[ \t]*end") (fgscript-previous-indentation-dedent))
     ((looking-at "^[ \t]*\\(fn\\|if\\|else\\|while\\|for\\|do\\)") (fgscript-previous-indentation))
     ((fgscript-previous-line-is-block-start) (+ (fgscript-previous-indentation) fgscript-indent-offset))
     ((looking-at "^[ \t]*\\(var\\|set\\|print\\|println\\|scan\\|#\\)") (fgscript-previous-indentation))
     (t (fgscript-previous-indentation)))))

(defun fgscript-previous-line-is-block-start ()
  "Check if the previous line starts a block."
  (save-excursion
    (forward-line -1)
    (beginning-of-line)
    (looking-at "^[ \t]*\\(fn\\|if\\|else\\|while\\|for\\|do\\)")))

(defun fgscript-previous-indentation ()
  "Get the indentation level of the previous non-blank line."
  (save-excursion
    (forward-line -1)
    (while (and (not (bobp)) (looking-at "^[ \t]*$"))
      (forward-line -1))
    (current-indentation)))

(defun fgscript-previous-indentation-dedent ()
  "Get the indentation level of the previous non-blank line and decrease it."
  (max 0 (- (fgscript-previous-indentation) fgscript-indent-offset)))

(defun fgscript-newline-and-indent ()
  "Insert a newline and indent the new line."
  (interactive)
  (newline)
  (fgscript-indent-line))

(define-derived-mode fgscript-mode prog-mode "FGScript"
  "Major mode for editing FGScript files."
  (setq font-lock-defaults
        '((
           ("\\<\\(fn\\|if\\|else\\|while\\|do\\|for\\|backward\\|from\\|to\\|step\\|end\\|var\\|set\\|call\\|print\\|println\\|scan\\)\\>" . font-lock-keyword-face)
           ("\\<\\(Integer\\|Real\\|Bool\\|String\\|None\\)\\>" . font-lock-type-face)
           ("\\[\\(Integer\\|Real\\|Bool\\|String\\)\\]" . font-lock-type-face)
           ("\\(==\\|!=\\|>=\\|<=\\|>\\|<\\|\\+\\|-\\|\\*\\|/\\|%\\|&&\\|||\\|!\\)" . font-lock-operator-face)
           ("\\<var\\s-+\\([a-zA-Z_][a-zA-Z_0-9]*\\)" 1 font-lock-variable-name-face)
           ("\\<fn\\s-+\\([a-zA-Z_][a-zA-Z_0-9]*\\)" 1 font-lock-function-name-face)
           )))

  (setq comment-start "#"
        comment-start-skip "^[ \t]*#+")

  (setq fgscript-indent-offset 2)

  (define-key fgscript-mode-map (kbd "TAB") 'fgscript-indent-line)
  (define-key fgscript-mode-map (kbd "RET") 'fgscript-newline-and-indent)

  (modify-syntax-entry ?# "<" fgscript-mode-syntax-table)
  (modify-syntax-entry ?\n ">" fgscript-mode-syntax-table))

(provide 'fgscript-mode)
(add-to-list 'auto-mode-alist '("\\.fgscr\\'" . fgscript-mode))
