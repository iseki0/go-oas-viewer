# Third-party notices

项目原创代码采用 Apache-2.0，见 LICENSE；项目 NOTICE 保留上游 swagger-ui-dist 的 NOTICE 原文。第三方代码保持各自许可。LICENSE 和 NOTICE 同时随 web/assets 中的副本内嵌。

本库内嵌由以下前端软件构建的资源：

| Package | Version | License | Source |
| --- | --- | --- | --- |
| @scalar/api-reference | 1.75.0 | MIT | https://github.com/scalar/scalar |
| @stoplight/elements | 9.0.29 | Apache-2.0 | https://github.com/stoplightio/elements |
| swagger-ui-dist | 5.34.0 | Apache-2.0 | https://github.com/swagger-api/swagger-ui |

完整的构建依赖版权、许可文本及传递依赖信息由 Vite 在构建时生成于 web/assets/third-party-licenses.txt，随 go:embed 一起分发。请保留该文件和静态资源中已有的版权声明。Go 依赖见 go.mod 与 go.sum，其许可随各 module 的源代码分发。

上游文档：

- Scalar 配置：https://scalar.com/products/api-references/configuration
- Elements 配置：https://github.com/stoplightio/elements/blob/main/docs/getting-started/elements/elements-options.md
- Swagger UI 配置：https://swagger.io/docs/open-source-tools/swagger-ui/usage/configuration/
