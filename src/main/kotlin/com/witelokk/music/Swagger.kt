package com.witelokk.music

import com.witelokk.music.models.FailureResponse
import io.github.smiley4.ktorswaggerui.SwaggerUI
import io.github.smiley4.ktorswaggerui.data.AuthScheme
import io.github.smiley4.ktorswaggerui.data.AuthType
import io.github.smiley4.schemakenerator.core.annotations.Format
import io.github.smiley4.schemakenerator.core.data.AnnotationData
import io.github.smiley4.schemakenerator.core.data.PrimitiveTypeData
import io.github.smiley4.schemakenerator.core.data.TypeId
import io.github.smiley4.schemakenerator.serialization.processKotlinxSerialization
import io.github.smiley4.schemakenerator.swagger.compileReferencingRoot
import io.github.smiley4.schemakenerator.swagger.data.TitleType
import io.github.smiley4.schemakenerator.swagger.generateSwaggerSchema
import io.github.smiley4.schemakenerator.swagger.handleCoreAnnotations
import io.github.smiley4.schemakenerator.swagger.withTitle
import io.ktor.server.application.*
import org.joda.time.DateTime
import java.util.*

fun Application.configureSwagger() {
    install(SwaggerUI) {
        info {
            title = "Music API"
            version = "latest"
        }
        schemas {
            generator = { type ->
                type
                    .processKotlinxSerialization {
                        customProcessor("UUID") {
                            PrimitiveTypeData(
                                id = TypeId.build(UUID::class.qualifiedName!!),
                                qualifiedName = String::class.qualifiedName!!,
                                simpleName = UUID::class.simpleName!!,
                                annotations = mutableListOf(
                                    AnnotationData(
                                        name = Format::class.qualifiedName!!,
                                        values = mutableMapOf("format" to "uuid")
                                    )
                                )
                            )
                        }
                        customProcessor("DateTime") {
                            PrimitiveTypeData(
                                id = TypeId.build(DateTime::class.qualifiedName!!),
                                qualifiedName = String::class.qualifiedName!!,
                                simpleName = DateTime::class.simpleName!!,
                                annotations = mutableListOf(
                                    AnnotationData(
                                        name = Format::class.qualifiedName!!,
                                        values = mutableMapOf("format" to "date-time")
                                    )
                                )
                            )
                        }
                    }
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
