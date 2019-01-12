;; This file contains a number of assertions about our s-expression language.
;; These assertions are run as part of the normal Go test suits by the test case
;; TestLispFileAssertions.
;; 
;; Each assertion must result in boolean value, if it True, then the test passes.
;; If it is False the test fails.
;;
;;;;;;;;;;;;;;;;;;
;; Mathematics: ;;
;;;;;;;;;;;;;;;;;;
(= 2 (+ 1 1))			; Addition of integers
(= 3.3 (+ 1.1 2.2))             ; Addition of floats
(= 10.6 (+ 1 2.2 3 4.4))        ; Mixed Addition
(= -3 (+ -2 -1))		; Addition of negative integers
(= -4 (+ 5 -9))			; Addition of mixed sign integers
(= -12.6 (+ 22.4 -35))          ; Addition of mixed precison and mixed sign
(= 0 (- 1 1))			; Subtraction of integers
(= 1.1 (- 2.2 1.1))             ; Subtraction of floats
(= 1.0 (- 10.6 4.4 3 2.2))      ; Mixed subtraction
(= -2 (- 0 1 1))		; Integer subtraction with negative result
(= -2.2 (- 1.0 1.1 2.1))        ; Float subtraction with a negative result
(= -2.2 (- 1 1.1 2 0.1))        ; Mixed subtraction with a negative result
(= 4 (- -2 -6))			; Subtraction of negative integers
(= 6.0 (- -4.1 -10.1))		; Subtraction of negatives floats




