package com.witelokk.music.models

import com.witelokk.music.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.*

@Serializable
data class Song(
    @Serializable(UUIDSerializer::class) val id: UUID,
    val name: String,
    @SerialName("cover_url") val coverUrl: String?,
    @SerialName("is_favorite") val isFavorite: Boolean,
    @SerialName("duration_seconds") val durationSeconds: Int,
    val artists: List<AristSummary>,
    @SerialName("stream_url") val streamUrl: String,
)