package com.witelokk.music

import io.ktor.serialization.kotlinx.json.*
import io.ktor.server.application.*
import io.ktor.server.plugins.contentnegotiation.*
import kotlinx.serialization.KSerializer
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.PrimitiveSerialDescriptor
import kotlinx.serialization.encoding.Decoder
import org.joda.time.DateTime
import java.util.*

fun Application.configureSerialization() {
    install(ContentNegotiation) {
        json()
    }
}

object UUIDSerializer : KSerializer<UUID> {
    override val descriptor = PrimitiveSerialDescriptor("UUID", PrimitiveKind.STRING)

    override fun deserialize(decoder: Decoder): UUID {
        return UUID.fromString(decoder.decodeString())
    }

    override fun serialize(encoder: kotlinx.serialization.encoding.Encoder, value: UUID) {
        encoder.encodeString(value.toString())
    }
}

object DateTimeSerializer : KSerializer<DateTime> {
    override val descriptor = PrimitiveSerialDescriptor("DateTime", PrimitiveKind.STRING)

    override fun deserialize(decoder: Decoder): DateTime {
        return DateTime.parse(decoder.decodeString())
    }

    override fun serialize(encoder: kotlinx.serialization.encoding.Encoder, value: DateTime) {
        encoder.encodeString(value.toString())
    }
}
