/**
 * @file KnockoutTournament.cpp
 * @brief Реализация турнира олимпийской системы.
 */

#include "KnockoutTournament.h"

KnockoutTournament::KnockoutTournament(
    const std::string& name,
    int maxParticipants,
    bool thirdPlaceMatch
)
    : Tournament(name, maxParticipants),
      thirdPlaceMatch(thirdPlaceMatch)
{
}

bool KnockoutTournament::hasThirdPlaceMatch() const
{
    return thirdPlaceMatch;
}