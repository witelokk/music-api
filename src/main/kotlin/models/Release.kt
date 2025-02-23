package com.witelokk.models

import com.witelokk.DateTimeSerializer
import com.witelokk.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import org.joda.time.DateTime
import java.util.*

@Serializable
data class Release(
    @Serializable(UUIDSerializer::class) val id: UUID,
    val name: String,
    @SerialName("cover_url") val coverUrl: String?,
    val type: String,
    @SerialName("released_at") @Serializable(DateTimeSerializer::class) val releasedAt: DateTime,
    val songs: Songs,
    val artists: ShortArtists,
)
