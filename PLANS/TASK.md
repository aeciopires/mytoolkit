<!-- TOC -->

- [TASK](#task)
  - [Requisitos funcionais:](#requisitos-funcionais)
  - [Requisitos não funcionais:](#requisitos-não-funcionais)

<!-- TOC -->

# TASK

Leia o arquivo README.md e crie um arquivo PLANS/PLAN_FEATURE_NAME.md para cada funcionalidade a ser implementada. 

A aplicacao deve ser desenvolvida web usando Golang.

## Requisitos funcionais:

- Tema dark e light
- A aplicacao deve funcionar no modo web e no modo CLI
- O modo web deve ser o padrão, mas a aplicacao deve aceitar um parametro de entrada para iniciar no modo CLI
- No CLI, a aplicacao deve aceitar parametros de entrada e saida para cada funcionalidade e ter um help para cada funcionalidade e um help geral para a aplicacao
- Criar endpoint REST para cada funcionalidade
- Criar interface web responsiva para cada funcionalidade
- Criar path de health check para a aplicação
- Criar path de metrics para a aplicação sobre o uso de cada funcionalidade, quantidade de requisições, tempo de resposta, vezes em que foi utilizada, ranking de uso.

Site que pode ser utilizado como referencia:

https://10015.io/tools/json-tree-viewer
https://10015.io/tools/qr-code-generator


## Requisitos não funcionais:

- Criar testes unitários para cada funcionalidade
- Criar documentação para cada funcionalidade
- Criar dockerfile para a aplicação
- Criar docker-compose para a aplicação
- Criar helm chart para a aplicação. Pode se inspirar no helm chart do projeto https://gitlab.com/aeciopires/kube-pires/-/tree/master/helm-chart
- Criar um script makefile para build, run, test, check dependencies e deploy da aplicação
- Toda a documentação e comentários devem ser escritos em inglês
- Utilize as versões mais novas das tecnologias
- Utilize as melhores praticas de codificação, system design, clean code
- Atualize o arquivo README.md com instruções de uso, da arquitetura, componentes de software utilizados, workflow em mermaid da aplicação e de cada funcionalidade, estrutura de diretórios
- Documente todos os endpoints REST com exemplos de uso e respostas
- Documente todos os comandos CLI com exemplos de uso e respostas
- Documente todos os testes unitários com exemplos de uso e respostas
- Documente todos os scripts de build, run, test, check dependencies e deploy com exemplos de uso e respostas
- Documente todas as variáveis de ambiente utilizadas na aplicação
- Criar um arquivo de CHANGELOG.md
- Criar um arquivo CLAUDE.md com informações relevantes ao projeto
- Criar uma seção de ROADMAP.md com informações sobre o roadmap do projeto
- Criar skills de desenvolvimento para cada funcionalidade na pasta .skills
- Teste o deploy da aplicacao via helm no cluster kind-king-multinodes
- Criar o arquivo de CONTRIBUTING.md semelhante a https://raw.githubusercontent.com/aeciopires/learning-istio/refs/heads/main/CONTRIBUTING.md