package com.witelokk.models

import com.witelokk.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.UUID

@Serializable
data class Playlist(
    @Serializable(UUIDSerializer::class) val id: UUID,
    val name: String,
    @SerialName("cover_url") val coverUrl: String?,
    @SerialName("songs_count") val songsCount: Int,
    val songs: Songs,
)
