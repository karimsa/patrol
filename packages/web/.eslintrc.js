// Source: @karimsa/wiz

module.exports = {
	"extends": [
		"plugin:react/recommended",
	],
	"plugins": [
		"import",
		"promise",
		"standard",
		"prettier",
		"node",
	],
	"settings": {
		"react": {
			"version": "detect"
		}
	},
	"parser": "babel-eslint",
	"parserOptions": {
		"ecmaVersion": 2018,
		"ecmaFeatures": {
			"jsx": true
		},
		"sourceType": "module"
	},
	"globals": {
		"document": true,
		"navigator": true,
		"window": true,
		"process": true,
	},
	"rules": {
		"accessor-pairs": "error",
		"arrow-spacing": "off",
		"block-spacing": "off",
		"brace-style": "off",
		"camelcase": [
			"error",
			{
				"properties": "never"
			}
		],
		"comma-dangle": [
			"error",
			"always-multiline"
		],
		"comma-spacing": "off",
		"comma-style": "off",
		"constructor-super": "error",
		"curly": 0,
		"dot-location": "off",
		"eol-last": "off",
		"eqeqeq": [
			"error",
			"always",
			{
				"null": "ignore"
			}
		],
		"func-call-spacing": "off",
		"generator-star-spacing": "off",
		"handle-callback-err": [
			"error",
			"^(err|error)$"
		],
		"indent": "off",
		"key-spacing": "off",
		"keyword-spacing": "off",
		"new-cap": [
			"error",
			{
				"newIsCap": true,
				"capIsNew": false
			}
		],
		"new-parens": "off",
		"no-array-constructor": "error",
		"no-caller": "error",
		"no-class-assign": "error",
		"no-compare-neg-zero": "error",
		"no-cond-assign": "error",
		"no-const-assign": "error",
		"no-constant-condition": [
			"error",
			{
				"checkLoops": false
			}
		],
		"no-control-regex": "error",
		"no-debugger": "error",
		"no-delete-var": "error",
		"no-dupe-args": "error",
		"no-dupe-class-members": "error",
		"no-dupe-keys": "error",
		"no-duplicate-case": "error",
		"no-empty-character-class": "error",
		"no-empty-pattern": "error",
		"no-eval": "error",
		"no-ex-assign": "error",
		"no-extend-native": "error",
		"no-extra-bind": "error",
		"no-extra-boolean-cast": "error",
		"no-extra-parens": "off",
		"no-fallthrough": "error",
		"no-floating-decimal": "off",
		"no-func-assign": "error",
		"no-global-assign": "error",
		"no-implied-eval": "error",
		"no-inner-declarations": [
			"error",
			"functions"
		],
		"no-invalid-regexp": "error",
		"no-irregular-whitespace": "error",
		"no-iterator": "error",
		"no-label-var": "error",
		"no-labels": "off",
		"no-lone-blocks": "error",
		"no-mixed-operators": 0,
		"no-mixed-spaces-and-tabs": "off",
		"no-multi-spaces": "off",
		"no-multi-str": "error",
		"no-multiple-empty-lines": "off",
		"no-negated-in-lhs": "error",
		"no-new": "error",
		"no-new-func": "error",
		"no-new-object": "error",
		"no-new-require": "error",
		"no-new-symbol": "error",
		"no-new-wrappers": "error",
		"no-obj-calls": "error",
		"no-octal": "error",
		"no-octal-escape": "error",
		"no-path-concat": "error",
		"no-proto": "error",
		"no-redeclare": "error",
		"no-regex-spaces": "error",
		"no-return-assign": [
			"error",
			"except-parens"
		],
		"no-return-await": "error",
		"no-self-assign": "error",
		"no-self-compare": "error",
		"no-sequences": "error",
		"no-shadow-restricted-names": "error",
		"no-sparse-arrays": "error",
		"no-tabs": "off",
		"no-template-curly-in-string": "error",
		"no-this-before-super": "error",
		"no-throw-literal": "error",
		"no-trailing-spaces": "off",
		"no-undef": "error",
		"no-undef-init": "error",
		"no-unexpected-multiline": 0,
		"no-unmodified-loop-condition": "error",
		"no-unneeded-ternary": [
			"error",
			{
				"defaultAssignment": false
			}
		],
		"no-unreachable": "error",
		"no-unsafe-finally": "error",
		"no-unsafe-negation": "error",
		"no-unused-expressions": [
			"error",
			{
				"allowShortCircuit": true,
				"allowTernary": true,
				"allowTaggedTemplates": true
			}
		],
		"no-unused-vars": [
			"error",
			{
				"varsIgnorePattern": "^_+$"
			}
		],
		"no-use-before-define": [
			"error",
			{
				"functions": false,
				"classes": false,
				"variables": false
			}
		],
		"no-useless-call": "error",
		"no-useless-computed-key": "error",
		"no-useless-constructor": "error",
		"no-useless-escape": "error",
		"no-useless-rename": "error",
		"no-useless-return": "error",
		"no-whitespace-before-property": "off",
		"no-with": "error",
		"object-curly-spacing": "off",
		"object-property-newline": "off",
		"one-var": [
			"error",
			{
				"initialized": "never"
			}
		],
		"operator-linebreak": "off",
		"padded-blocks": "off",
		"prefer-promise-reject-errors": "error",
		"quotes": 0,
		"rest-spread-spacing": "off",
		"semi": "off",
		"semi-spacing": "off",
		"space-before-blocks": "off",
		"space-before-function-paren": "off",
		"space-in-parens": "off",
		"space-infix-ops": "off",
		"space-unary-ops": "off",
		"spaced-comment": [
			"error",
			"always",
			{
				"line": {
					"markers": [
						"*package",
						"!",
						"/",
						",",
						"="
					]
				},
				"block": {
					"balanced": true,
					"markers": [
						"*package",
						"!",
						",",
						":",
						"::",
						"flow-include"
					],
					"exceptions": [
						"*"
					]
				}
			}
		],
		"symbol-description": "error",
		"template-curly-spacing": "off",
		"template-tag-spacing": "off",
		"unicode-bom": "off",
		"use-isnan": "error",
		"valid-typeof": "error",
		"wrap-iife": "off",
		"yield-star-spacing": "off",
		"yoda": [
			"error",
			"never"
		],
		"import/export": 2,
		"import/first": "error",
		"import/no-duplicates": "error",
		"import/no-named-default": "error",
		"import/no-webpack-loader-syntax": "error",
		"node/no-deprecated-api": "off",
		"node/process-exit-as-throw": "error",
		"promise/param-names": "off",
		"standard/array-bracket-even-spacing": "off",
		"standard/computed-property-even-spacing": "off",
		"standard/no-callback-literal": "error",
		"standard/object-curly-even-spacing": "off",
		"import/no-unresolved": "off",
		"import/named": "off",
		"import/namespace": "off",
		"import/default": 2,
		"arrow-body-style": 0,
		"lines-around-comment": 0,
		"max-len": 0,
		"no-confusing-arrow": 0,
		"prefer-arrow-callback": 0,
		"array-bracket-newline": "off",
		"array-bracket-spacing": "off",
		"array-element-newline": "off",
		"arrow-parens": "off",
		"computed-property-spacing": "off",
		"function-paren-newline": "off",
		"generator-star": "off",
		"implicit-arrow-linebreak": "off",
		"indent-legacy": "off",
		"linebreak-style": "off",
		"multiline-ternary": "off",
		"newline-per-chained-call": "off",
		"no-arrow-condition": "off",
		"no-comma-dangle": "off",
		"no-extra-semi": "off",
		"no-reserved-keys": "off",
		"no-space-before-semi": "off",
		"no-spaced-func": "off",
		"no-wrap-func": "off",
		"nonblock-statement-body-position": "off",
		"object-curly-newline": "off",
		"one-var-declaration-per-line": "off",
		"quote-props": [
			"error",
			"as-needed"
		],
		"semi-style": "off",
		"space-after-function-name": "off",
		"space-after-keywords": "off",
		"space-before-function-parentheses": "off",
		"space-before-keywords": "off",
		"space-in-brackets": "off",
		"space-return-throw-case": "off",
		"space-unary-word-ops": "off",
		"switch-colon-spacing": "off",
		"wrap-regex": "off",
		"prettier/prettier": [
			"error",
			{
				"singleQuote": true,
				"semi": false,
				"trailingComma": "all",
				"useTabs": true
			},
			{
				"usePrettierrc": false
			}
		],
		"import/order": [
			"error",
			{
				"groups": [
					"builtin",
					"external",
					"internal"
				],
				"newlines-between": "always"
			}
		],
		"no-implicit-coercion": "error"
	},
	"env": {
		"browser": true,
		"es6": true,
	}
}
