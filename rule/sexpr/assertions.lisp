;; This file contains a number of assertions about our s-expression language.
;; These assertions are run as part of the normal Go test suits by the test case
;; TestLispFileAssertions in parser_internal_test.go.
;; 
;; Each assertion must result in boolean value, if it True, then the test passes.
;; If it is False the test fails.
;;
;;;;;;;;;;;;;;;;;;
;; Mathematics: ;;
;;;;;;;;;;;;;;;;;;
(assert= 2 (+ 1 1))			; Addition of integers
(assert= 3.3 (+ 1.1 2.2))		; Addition of floats
(assert= 10.6 (+ 1 2.2 3 4.4))		; Mixed Addition
(assert= -3 (+ -2 -1))			; Addition of negative integers
(assert= -4 (+ 5 -9))			; Addition of mixed sign integers
(assert= -12.6 (+ 22.4 -35))		; Addition of mixed precision and mixed sign
(assert= 0 (- 1 1))			; Subtraction of integers
(assert= 1.1 (- 2.2 1.1))		; Subtraction of floats
(assert= 1.0 (- 10.6 4.4 3 2.2))	; Mixed subtraction
(assert= -2 (- 0 1 1))			; Integer subtraction with negative result
(assert= -2.2 (- 1.0 1.1 2.1))		; Float subtraction with a negative result
(assert= -2.2 (- 1 1.1 2 0.1))		; Mixed subtraction with a negative result
(assert= 4 (- -2 -6))			; Subtraction of negative integers
(assert= 6.0 (- -4.1 -10.1))		; Subtraction of negatives floats
(assert= 99 (* 11 9))			; Multiplication of integers
(assert= 146.8944 (* 12.12 12.12))	; Multiplication of floats
(assert= 220 (* -22 -10))		; Multiplication of negative integers
(assert= 475.6 (* -20.5 -23.2))		; Multiplication of negative floats
(assert= -220 (* 22 -10))		; Multiplication of mixed sign integers
(assert= -475.6 (* 20.5 -23.2))		; Multiplication of mixed sign floats
(assert= 1 (/ 10 10))			; Integer division
(assert= 2.0 (/ 2.2 1.1))		; Float division
(assert= 1 (/ -10 -10 1))		; Division of negative integers
(assert= -1 (/ 10 -10 1))		; Integer division with mixed signs
(assert= 1.0 (/ -2.2 -1.1 -1.0 -2.0))	; Division of negative floats
(assert= -2.0 (/ 2.2 -1.1))		; Float division with mixed signs
(assert=  2 (% 5 3))			; Modulo of integers
(assert= -2 (% -5 3))			; Modulo of mixed sign integers (negative dividend)
(assert= 2 (% 5 -3))			; Modulo of mixed sign integers (negative divisor)
(assert= -2 (% -5 -3))			; Modulo of negative integers

;;;;;;;;;;;;;;;;;;;;;
;; Cast operations ;;
;;;;;;;;;;;;;;;;;;;;;
(assert= 1.0 (int->float 1))		; Cast positive integer to float
(assert= -12.0 (int->float -12))	; Cast negative integer to float
