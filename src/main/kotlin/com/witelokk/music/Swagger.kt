package com.witelokk.music

import com.witelokk.music.models.FailureResponse
import io.github.smiley4.ktoropenapi.OpenApi
import io.github.smiley4.ktoropenapi.config.AuthScheme
import io.github.smiley4.ktoropenapi.config.AuthType
import io.github.smiley4.schemakenerator.serialization.SerializationSteps.analyzeTypeUsingKotlinxSerialization
import io.github.smiley4.schemakenerator.swagger.SwaggerSteps.compileReferencingRoot
import io.github.smiley4.schemakenerator.swagger.SwaggerSteps.generateSwaggerSchema
import io.github.smiley4.schemakenerator.swagger.SwaggerSteps.handleCoreAnnotations
import io.github.smiley4.schemakenerator.swagger.SwaggerSteps.withTitle
import io.github.smiley4.schemakenerator.swagger.data.TitleType
import io.ktor.server.application.*

fun Application.configureSwagger() {
    install(OpenApi) {
        info {
            title = "Music API"
            version = "latest"
        }
        schemas {
            generator = { type ->
                type
                    .analyzeTypeUsingKotlinxSerialization()
//                    .processKotlinxSerialization {
//                        customProcessor("UUID") {
//                            PrimitiveTypeData(
//                                id = TypeId.build(UUID::class.qualifiedName!!),
//                                qualifiedName = String::class.qualifiedName!!,
//                                simpleName = UUID::class.simpleName!!,
//                                annotations = mutableListOf(
//                                    AnnotationData(
//                                        name = Format::class.qualifiedName!!,
//                                        values = mutableMapOf("format" to "uuid")
//                                    )
//                                )
//                            )
//                        }
//                        customProcessor("DateTime") {
//                            PrimitiveTypeData(
//                                id = TypeId.build(DateTime::class.qualifiedName!!),
//                                qualifiedName = String::class.qualifiedName!!,
//                                simpleName = DateTime::class.simpleName!!,
//                                annotations = mutableListOf(
//                                    AnnotationData(
//                                        name = Format::class.qualifiedName!!,
//                                        values = mutableMapOf("format" to "date-time")
//                                    )
//                                )
//                            )
//                        }
//                        markNotParameterized<ShortArtist>()
//                    }
                    .generateSwaggerSchema()
                    .handleCoreAnnotations()
                    .withTitle(TitleType.OPENAPI_SIMPLE)
                    .compileReferencingRoot()
            }
        }
        security {
            securityScheme("Authorization") {
                type = AuthType.HTTP
                scheme = AuthScheme.BEARER
            }
            defaultSecuritySchemeNames("Authorization")
            defaultUnauthorizedResponse {
                description = "Token is not valid or has expired"
                body<FailureResponse>()
            }

        }
    }
}
