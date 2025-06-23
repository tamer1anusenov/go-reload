# go-reloaded

Welcome to **go-reloaded** — a simple yet powerful text transformation and correction tool written in Go.

## 🧠 Project Overview

This project serves as a refresher on Go fundamentals by reusing and extending your past code. You'll build a command-line tool that reads a text file, performs a series of transformations, and writes the corrected output to another file.

> ✅ This project will be reviewed by your peers, and you’ll also act as a reviewer for others. Test your code thoroughly!

---

## 🎯 Objectives

- Create a CLI tool to process and correct text based on specific rules.
- Practice working with the Go filesystem and text manipulation.
- Follow Go best practices and include unit tests for your functions.

---

## 🔧 Features

The tool modifies input text according to the following rules:

### 🔢 Number Transformations

- `(hex)` – Converts the previous **hexadecimal** word to decimal.  
  Example: `1E (hex)` → `30`

- `(bin)` – Converts the previous **binary** word to decimal.  
  Example: `10 (bin)` → `2`

### 🔤 Casing Transformations

- `(up)` – Converts the previous word to **UPPERCASE**.
- `(low)` – Converts the previous word to **lowercase**.
- `(cap)` – Converts the previous word to **Capitalized** (first letter uppercase).

You can apply transformations to multiple words:  
Examples:  
- `(up, 2)` → converts two previous words to uppercase.  
- `(cap, 3)` → capitalizes three previous words.

### 🔡 Articles

- Replace `a` with `an` if the next word begins with a vowel or 'h'.  
  Example: `a amazing rock` → `an amazing rock`

### ✨ Punctuation Formatting

- Ensure that punctuation marks (`.`, `,`, `!`, `?`, `:`, `;`) stick to the **previous** word with **no space**, and are **spaced from the next word** unless grouped (`...`, `!?`, etc.).
  
- Handle quotes `'` correctly:  
  - `' word '` → `'word'`  
  - `' multiple words '` → `'multiple words'`

---

## 🛠️ Usage

```sh
$ cat sample.txt
it (cap) was the best of times, it was the worst of times (up) , it was the age of wisdom...

$ go run . sample.txt result.txt

$ cat result.txt
It was the best of times, it was the worst of TIMES, it was the age of wisdom...

##to run this 

go run . <input.txt> <output.txt>
