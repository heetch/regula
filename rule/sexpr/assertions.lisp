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
(assert= 1 (float->int 1.5))		; Cast a positive float to an integer
(assert= -13 (float->int -13.3939))	; Cast a negative float to an integer


;;;;;;;;;;;;;;;;;;;;;;;;
;; Control operations ;;
;;;;;;;;;;;;;;;;;;;;;;;;
(assert= 10 (let x 10 x))				; Return bound value from let
(assert= 10 (let x (+ 5 5) x))				; Let identifier equal result of value form
(assert= #true (let f #false (not f)))			; Operate on bound value
(assert= #true (if #true #true #false))			; If returns the true-part when the condition is true
(assert= #false (if #false #true #false))		; If returns the false-part when the condition is false
(assert= #true (if (= 2 (+ 1 1)) #true #false))		; If tests can be nested expressions
(assert= #true (if #true (= 2 (+ 1 1)) #false))		; If true-parts can be nested expressions
(assert= #false (if #false #true (= 3 (+ 1 1))))	; If false-parts can be nested expressions

;;;;;;;;;;;;;;;;;;;;;;;;;;;;
;; Equality and compasion ;;
;;;;;;;;;;;;;;;;;;;;;;;;;;;;
(assert= #true (= 1 1))						; Integer equaliyt
(assert= #false (= 1 2))					; Integer inequality
(assert= #true (= 1.1 1.1))					; Float equality
(assert= #false (= 1.1 1.2))					; Float inequality
(assert= #true (= "Foo" "Foo"))					; String equality
(assert= #false (= "Foo" "foo"))				; String inequality (by case)
(assert= #false (= "foo" "bar"))				; String inequality
(assert= #true (= #true #true))					; Boolean equality
(assert= #false (= #true #false))				; Boolean inequality
(assert= #true (< 0 1))						; integer less than (true)
(assert= #false (< 0 0))					; integer less than (false because equal)
(assert= #false (< 0 -1))					; integer less than (false because greater than)
(assert= #true (< 0.1 0.2))					; integer less than (true)
(assert= #false (< 0.1 0.1))					; integer less than (false because equal)
(assert= #false (< 0.1 0))					; integer less than (false because greater than)
(assert= #true (< "Abba" "BeeGees"))				; String less than (ASCIIbetical)
(assert= #true (< "Eagle" "Eagles"))				; String less than (Length)
(assert= #true (< "A" "a"))					; String less than (Case)
(assert= #true (< "A" "a" "b"))					; Multi-string less than
(assert= #false (< "A" "A"))					; Identical strings, therefore no inequality
(assert= #false (< "AB" "A"))					; Second string is less than the first, therefore false
(assert= #false (< "A" "a" "A"))				; Multi-string false case (last string is >)
(assert= #true (< #false #true))				; Boolean, False < True = True
(assert= #false (< #false #false))				; Boolean, False < False = False
(assert= #false (< #true #true))				; Boolean, True < True = False
(assert= #false (< #true #false))				; Boolean, True < False = False
(assert= #true (> 1 0))						; integer greater than (true)
(assert= #false (> 0 0))					; integer greater than (false because equal)
(assert= #false (> -1 0))					; integer greater than (false because greater than)
(assert= #true (> 0.2 0.1))					; integer greater than (true)
(assert= #false (> 0.1 0.1))					; integer greater than (false because equal)
(assert= #false (> 0 0.1))					; integer greater than (false because greater than)
(assert= #true (> "ZZ Top" "BeeGees"))				; String greater than (ASCIIbetical)
(assert= #true (> "Eagles" "Eagle"))				; String greater than (Length)
(assert= #true (> "a" "A"))					; String greater than (Case)
(assert= #true (> "b" "a" "A"))					; Multi-string greater than
(assert= #false (> "A" "A"))					; Identical strings, therefore no inequality
(assert= #false (> "A" "AB"))					; Second string is greater than the first, therefore false
(assert= #false (> "A" "a" "A"))				; Multi-string false case (last string is >)
(assert= #true (> #true #false))				; Boolean, False > True = True
(assert= #false (> #false #false))				; Boolean, False > False = False
(assert= #false (> #true #true))				; Boolean, True > True = False
(assert= #false (> #false #true))				; Boolean, True > False = False
(assert= #true (<= 2 3))					; Integer <= (true because less than)
(assert= #true (<= 3 3))					; Integer <= (true because equal)
(assert= #true (<= 2 3 3))					; Integer <= (true because a < b = c)
(assert= #true (<= 3 3 4))					; Integer <= (true because a = b < c)
(assert= #true (<= 2 3 4))					; Integer <= (true because a < b < c)
(assert= #true (<= 2 2 2))					; Interger <= (true because a = b = c)
(assert= #false (<= 3 2 2))					; Interger <= (false because a > b = c)
(assert= #false (<= 2 3 2))					; Interger <= (false because a < b > c)
(assert= #false (<= 3 3 2))					; Interger <= (false because a = b > c)
(assert= #true (>= 1.3 1.2))					; Float >= (true because greater than)
(assert= #true (>= 1.3 1.3))					; Float >= (true because equal)
(assert= #true (>= 1.3 1.3 1.2))				; Float >= (true because a = b > c)
(assert= #true (>= 1.3 1.2 1.2))				; Float >= (true because a > b = c)
(assert= #true (>= 1.3 1.3 1.3))				; Float >= (true because a = b = c)
(assert= #false (>= 1.2 1.3))					; Float >= (false because a < b)
(assert= #false (>= 1.2 1.2 1.3))				; Float >= (false because a = b < c)
(assert= #false (>= 1.1 1.2 1.3))				; Float >= (false because a < b < c)
(assert= #false (>= 1.3 1.2 1.3))				; Float >= (false because a > b < c)
(assert= #false (>= 1.1 1.2 1.2))				; Float >= (false because a < b = c)
(assert= #false (>= 1.2 1.3 1.2))				; Float >= (false because a < b > c) 
(assert= #true (<= 1.2 1.3))					; Float <= (true because less than)
(assert= #true (<= 1.3 1.3))					; Float <= (true because equal)
(assert= #true (<= 1.2 1.3 1.3))				; Float <= (true because a < b = c)
(assert= #true (<= 1.3 1.3 4))					; Float <= (true because a = b < c)
(assert= #true (<= 1.2 1.3 4))					; Float <= (true because a < b < c)
(assert= #true (<= 1.2 1.2 1.2))				; Interger <= (true because a = b = c)
(assert= #false (<= 1.3 1.2 1.2))				; Interger <= (false because a > b = c)
(assert= #false (<= 1.2 1.3 1.2))				; Interger <= (false because a < b > c)
(assert= #false (<= 1.3 1.3 1.2))				; Interger <= (false because a = b > c)
(assert= #true (<= "ABBA" "Beegees"))				; String <= (true because a < b)
(assert= #true (<= "ABBA" "ABBA"))				; String <= (true because a = b)
(assert= #true (<= "ABBA" "Beegees" "Chumbawumba"))		; String <= (true because a < b < c)
(assert= #true (<= "ABBA" "Beegees" "Beegees"))			; String <= (true because a < b = c)
(assert= #true (<= "Beegees" "Beegees" "Beegees"))		; String <= (true because a = b = c)
(assert= #false (<= "Beegees" "ABBA"))				; String <= (false because a > b)
(assert= #false (<= "Beegees" "ABBA" "ABBA"))			; String <= (false because a > b = c)
(assert= #false (<= "ABBA" "Beegees" "ABBA"))			; String <= (false because a < b > c)
(assert= #false (<= "Beegees" "Beegees" "ABBA"))		; String <= (false because a = b > c)
(assert= #true (<= #false #true))				; Boolean <= (true because a < b)
(assert= #true (<= #true #true))				; Boolean <= (true because a = b)
(assert= #true (<= #false #false #false))			; Boolean <= (true because a = b = c)
(assert= #true (<= #false #true #true))				; Boolean <= (true because a < b = c)
(assert= #true (<= #false #false #true))			; Boolean <= (true because a = b < c)
(assert= #false (<= #true #false))				; Boolean <= (false because a > b)
(assert= #false (<= #true #true #false))			; Boolean <= (false because a = b > c)
(assert= #false (<= #true #false #false))			; Boolean <= (false because a > b = c)
(assert= #true (>= 1.3 1.2))					; Float >= (true because greater than)
(assert= #true (>= 1.3 1.3))					; Float >= (true because equal)
(assert= #true (>= 1.3 1.3 1.2))				; Float >= (true because a = b > c)
(assert= #true (>= 1.3 1.2 1.2))				; Float >= (true because a > b = c)
(assert= #true (>= 1.3 1.3 1.3))				; Float >= (true because a = b = c)
(assert= #false (>= 1.2 1.3))					; Float >= (false because a < b)
(assert= #false (>= 1.2 1.2 1.3))				; Float >= (false because a = b < c)
(assert= #false (>= 1.1 1.2 1.3))				; Float >= (false because a < b < c)
(assert= #false (>= 1.3 1.2 1.3))				; Float >= (false because a > b < c)
(assert= #false (>= 1.1 1.2 1.2))				; Float >= (false because a < b = c)
(assert= #false (>= 1.2 1.3 1.2))				; Float >= (false because a < b > c) 
(assert= #true (>= "ZZ-Top" "Uriah Heap"))			; String >= (true because a > b)
(assert= #true (>= "ZZ-Top" "ZZ-Top"))				; String >= (true because a = b)
(assert= #true (>= "ZZ-Top" "Uriah Heap" "T-Rex"))		; String >= (true because a > b > c)
(assert= #true (>= "ZZ-Top" "Uriah Heap" "Uriah Heap"))		; String >= (true because a > b = c)
(assert= #true (>= "Uriah Heap" "Uriah Heap" "Uriah Heap"))	; String >= (true because a = b = c)
(assert= #false (>= "Uriah Heap" "ZZ-Top"))			; String >= (false because a > b)
(assert= #false (>= "Uriah Heap" "ZZ-Top" "ZZ-Top"))		; String >= (false because a > b = c)
(assert= #false (>= "ZZ-Top" "Uriah Heap" "ZZ-Top"))		; String >= (false because a > b > c)
(assert= #false (>= "Uriah Heap" "Uriah Heap" "ZZ-Top"))	; String >= (false because a = b > c)
(assert= #true (>= #true #false))				; Boolean >= (true because a > b)
(assert= #true (>= #false #false))				; Boolean >= (true because a = b)
(assert= #true (>= #true #true #true))				; Boolean >= (true because a = b = c)
(assert= #true (>= #true #false #false))			; Boolean >= (true because a > b = c)
(assert= #true (>= #true #true #false))				; Boolean >= (true because a = b > c)
(assert= #false (>= #false #true))				; Boolean >= (false because a < b)
(assert= #false (>= #false #false #true))			; Boolean >= (false because a = b < c)
(assert= #false (>= #false #true #true))			; Boolean >= (false because a < b = c)

;;;;;;;;;;;;;
;; Hashing ;;
;;;;;;;;;;;;;
(assert= 2179869525 (fnv 1234))			;  FNV Hash of an integer
(assert= 566939793 (fnv 1234.1234))		;  FNV Hash of a Float64
(assert= 536463009 (fnv "travelling in style"))	;  FNV Hash of a string
(assert= 3053630529 (fnv #true))		;  FNV Hash of a boolean (true)
(assert= 2452206122 (fnv #false))		;  FNV Hash of a boolean (false)


;;;;;;;;;;;;;;
;; Grouping ;;
;;;;;;;;;;;;;;
(assert= #true (percentile "Bob Dylan" 96)) ; Check that "Bob Dylan" is in the the 96th percentile.
(assert= #false (percentile "Joni Mitchell" 96)) ; Check that "Joni Mitchell" is not in the the 96th percentile.
