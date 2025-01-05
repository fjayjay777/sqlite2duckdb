LEXER_FILE ?= SQLiteLexer.g4
PARSER_FILE ?= SQLiteParser.g4
OUTPUT_DIR ?= internal/parser

clean-parser:
	rm -rf $(OUTPUT_DIR)/*

gen-lexer: clean-parser 
	antlr4 -Dlanguage=Go -visitor -o $(OUTPUT_DIR) $(LEXER_FILE)

gen-parser: gen-lexer
	antlr4 -Dlanguage=Go -visitor -o $(OUTPUT_DIR) $(PARSER_FILE)

run:
	go run ./internal/main.go

.PHONY: gen-lexer gen-parser clean-parser
