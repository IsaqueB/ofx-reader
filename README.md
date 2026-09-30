# ofxgo

Biblioteca Go, sem dependências externas, para interpretação de arquivos OFX.

Nesta primeira versão, o projeto implementa **OFX 1.0.2 (`VERSION:102`)**, baseado em SGML, com foco em extratos bancários (`BANKMSGSRSV1` / `STMTRS`). A arquitetura já possui pontos de extensão para outras versões OFX e para perfis específicos de bancos.

## Objetivos

- Go puro / somente standard library.
- Ler OFX 1.0.2 real, inclusive tags folha sem fechamento explícito.
- Preservar valores monetários sem conversão automática para `float64`.
- Suportar headers e charsets comuns em arquivos brasileiros, inclusive Windows-1252.
- Fornecer API estável e independente do banco.
- Permitir adicionar OFX 2.x posteriormente sem alterar os consumidores da lib.
- Permitir adicionar profiles de Banco do Nordeste, Banco do Brasil, Bradesco, Itaú e Inter sem alterar o parser genérico.

## Instalação

```bash
go get github.com/bracomil/ofxgo
```

> Enquanto o módulo não estiver publicado nesse endereço, use `replace` no `go.mod` ou altere o module path para o repositório onde você hospedar o projeto.

## Uso básico

```go
package main

import (
    "fmt"
    "log"

    "github.com/bracomil/ofxgo/ofx"
)

func main() {
    doc, err := ofx.ParseFile("extrato.ofx")
    if err != nil {
        log.Fatal(err)
    }

    for _, statement := range doc.Statements {
        fmt.Println("Banco:", statement.Account.BankID)
        fmt.Println("Agência:", statement.Account.BranchID)
        fmt.Println("Conta:", statement.Account.AccountID)

        for _, tx := range statement.Transactions {
            fmt.Printf("%s | %s | %s | %s\n",
                tx.PostedAt.Format("2006-01-02"),
                tx.Amount,
                tx.FITID,
                tx.Memo,
            )
        }
    }
}
```

Também existem:

```go
ofx.Parse(reader)
ofx.ParseBytes(data)
ofx.ParseFile(path)
```

Para uma instância reutilizável:

```go
parser := ofx.NewParser()
doc, err := parser.ParseFile("extrato.ofx")
```

## Valores monetários

`Transaction.Amount` e `Balance.Amount` são strings decimais exatas:

```go
fmt.Println(tx.Amount) // "1234.56"
```

Para operações matemáticas exatas sem dependências externas:

```go
amount, err := tx.AmountRat() // *big.Rat
```

Isso evita introduzir erros binários de `float64` na camada de parsing/conciliação.

## Extensão para novas versões OFX

O parser usa a interface:

```go
type VersionParser interface {
    Version() string
    Parse(header Header, body []byte) (*Document, error)
}
```

Uma implementação futura para OFX 2.x pode ser registrada com:

```go
parser.RegisterVersion(myOFX2Parser)
```

Os consumidores continuam recebendo o mesmo `ofx.Document`.

## Profiles de bancos

Profiles ainda não contêm regras específicas. A infraestrutura já está pronta:

```go
type BankProfile interface {
    Name() string
    Match(header Header, statement BankStatement) bool
    Normalize(statement *BankStatement) error
}
```

Depois poderemos implementar, por exemplo:

- `BNBProfile`
- `BancoDoBrasilProfile`
- `BradescoProfile`
- `ItauProfile`
- `InterProfile`

sem modificar o parser OFX 1.0.2.

## Estrutura

```text
ofxgo/
├── go.mod
├── README.md
├── ofx/
│   ├── date.go
│   ├── parser.go
│   ├── profile.go
│   ├── types.go
│   ├── v102.go
│   └── *_test.go
├── internal/
│   ├── sgml/
│   │   ├── node.go
│   │   └── parser.go
│   └── textcodec/
│       └── decode.go
└── testdata/
    └── sample_102.ofx
```

## Escopo da versão atual

Implementado:

- Header OFX 1.x.
- `VERSION:102`.
- SGML com fechamento implícito de elementos folha.
- `BANKACCTFROM`.
- `BANKTRANLIST`.
- `STMTTRN`.
- `LEDGERBAL`.
- `AVAILBAL`.
- Datas OFX com timezone e frações de segundo.
- Windows-1252, ISO-8859-1 e UTF-8.
- Campos desconhecidos simples preservados em `Raw`.

Ainda não implementado:

- OFX 2.x/XML.
- Cartões de crédito.
- Investimentos.
- Loans.
- Profiles específicos por banco.
- Validação integral contra DTD.

## Referência do padrão

A Financial Data Exchange mantém o padrão OFX e lista o OFX 1.0.2 entre as versões anteriores. O OFX 1.x utiliza sintaxe SGML; OFX 2.x migrou para XML.

https://financialdataexchange.org/about-fdx/ofx-work-group/
