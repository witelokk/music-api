package com.witelokk.music.models

import com.witelokk.music.UUIDSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.UUID

@Serializable
data class AristSummary(
    @Serializable(UUIDSerializer::class) val id: UUID,
    val name: String,
    @SerialName("avatar_url") val avatarUrl: String?,
)