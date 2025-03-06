package com.witelokk.music.models

import com.witelokk.music.DateTimeSerializer
import com.witelokk.music.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import org.joda.time.DateTime
import java.util.*

@Serializable
data class ShortRelease(
    @Serializable(UUIDSerializer::class) val id: UUID,
    val name: String,
    @SerialName("cover_url") val coverUrl: String?,
    val type: String,
    @SerialName("released_at") @Serializable(DateTimeSerializer::class) val releasedAt: DateTime,
    val artists: ShortArtists,
)
