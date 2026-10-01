#pragma once

#include "Tournament.h"

/**
 * @file KnockoutTournament.h
 * @brief Объявление турнира олимпийской системы.
 */

/**
 * @brief Турнир олимпийской системы.
 *
 * Класс расширяет базовый класс Tournament
 * и представляет соревнование на выбывание.
 *
 * Дополнительно класс хранит информацию
 * о необходимости проведения матча за третье место.
 */
class KnockoutTournament : public Tournament
{
public:
    /**
     * @brief Создаёт турнир олимпийской системы.
     *
     * @param name Название турнира.
     * @param maxParticipants Максимальное количество участников.
     * @param thirdPlaceMatch Необходимость проведения матча за третье место.
     */
    KnockoutTournament(
        const std::string& name,
        int maxParticipants,
        bool thirdPlaceMatch
    );

    /**
     * @brief Проверяет наличие матча за третье место.
     *
     * @return true, если матч за третье место проводится;
     * false в противном случае.
     */
    bool hasThirdPlaceMatch() const;

private:
    bool thirdPlaceMatch; ///< Признак проведения матча за третье место.
};