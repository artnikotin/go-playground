package search_v3

import (
	"fmt"
	"github.com/KosyanMedia/delta/pkg/currency"
	"github.com/KosyanMedia/delta/pkg/iata"
	"github.com/KosyanMedia/delta/pkg/types/datetime"
	"github.com/KosyanMedia/delta/pkg/types/search/base"
	"github.com/KosyanMedia/delta/pkg/types/search/delta"
	v3 "github.com/KosyanMedia/delta/search/cmd/results-api/api/v3"
	"github.com/KosyanMedia/delta/search/cmd/results-api/filter"
	"github.com/KosyanMedia/delta/search/cmd/results-api/filter/boundaries"
	"github.com/KosyanMedia/delta/search/cmd/results-api/filter/times"
	"github.com/KosyanMedia/delta/search/cmd/results-api/filter/transfers"
	"golang.org/x/exp/constraints"
)

func resultsToProto(results v3.SearchResults) *SearchResults {
	return &SearchResults{
		Chunks: chunksToProto(results),
	}
}

func chunksToProto(chunks []*v3.Chunk) []*Chunk {
	result := make([]*Chunk, len(chunks))
	for i, chunk := range chunks {
		paymentOptions := make([]string, len(chunk.SearchParams.PaymentOptions))
		for i2, option := range chunk.SearchParams.PaymentOptions {
			paymentOptions[i2] = string(option)
		}

		result[i] = &Chunk{
			ChunkId:                              chunk.ChunkID,
			LastUpdateTimestamp:                  chunk.LastUpdateTimestamp,
			DebugInfo:                            debugInfoToProto(chunk.DebugInfo),
			Tickets:                              convArray(chunk.Tickets, ticketToProto),
			SoftTickets:                          softResponseToProto(chunk.SoftTickets),
			BrandTicket:                          ticketToProtoOpt(chunk.BrandTicket),
			BrandTickets:                         convMap(chunk.BrandTickets, convNum[int, int64], ticketToProto),
			CheapestTicket:                       ticketToProtoOpt(chunk.CheapestTicket),
			FilteredCheapestTicket:               ticketToProtoOpt(chunk.FilteredCheapestTicket),
			CheapestTicketWithoutAirportPrecheck: ticketToProtoOpt(chunk.CheapestTicketWithoutAirportPreCheck),
			DirectFlights:                        convArray(chunk.DirectFlights, directFlightsToProto),
			FlightLegs:                           convArray(chunk.FlightLegs, flightLegToProto),
			Airlines:                             convMap(chunk.Airlines, convString[iata.AirlineID], airlineInfoToProto),
			Places:                               placesToProto(chunk.Places),
			Agents:                               convMap(chunk.Agents, convNum[int, int64], agentInfoToProto),
			Alliances:                            convMap(chunk.Alliances, convNum[int, int64], allianceToProto),
			Equipments:                           convMap(chunk.Equipments, same[string], equipmentToProto),
			SearchParams: &SearchParams{
				Passengers: &Passengers{
					Adults:   chunk.SearchParams.Passengers.Adults,
					Children: chunk.SearchParams.Passengers.Children,
					Infants:  chunk.SearchParams.Passengers.Infants,
				},
				TripClass:      TripClass(chunk.SearchParams.TripClass),
				SourceKind:     SourceKind(chunk.SearchParams.SourceKind),
				Experiments:    chunk.SearchParams.Experiments,
				PaymentOptions: paymentOptions,
			},
			DegradedFilterBoundaries: degradedBoundariesToProto(chunk.DegradedFilterBoundaries),
			FilterBoundaries:         boundariesToProto(chunk.FilterBoundaries),
			Meta: &ResultsMeta{
				FilteredTicketsCount: int64(chunk.Meta.FilteredTicketsCount),
				TotalTicketsCount:    int64(chunk.Meta.TotalTicketsCount),
				DirectTicketsCount:   int64(chunk.Meta.DirectTicketsCount),
			},
			FilterState: filterState(chunk.FilterState),
			Order:       Order(chunk.Order),
			Brand:       Brand(chunk.Brand),
		}
	}
	return result
}

func debugInfoToProto(info *v3.DebugInfo) *DebugInfo {
	if info == nil {
		return nil
	}
	return &DebugInfo{
		ServerName:      info.ServerName,
		DataCenter:      info.DataCenter,
		Gates:           gatesToProto(info.Gates),
		FromCache:       info.FromCache,
		SearchStartTime: info.SearchStartTime.UnixMilli(),
	}
}

func gatesToProto(gates map[v3.GateName]v3.GateDebugInfo) map[string]*GateDebugInfo {
	if gates == nil {
		return nil
	}
	result := make(map[string]*GateDebugInfo, len(gates))
	for name, info := range gates {
		result[name] = &GateDebugInfo{
			Name:                    info.Name,
			Agents:                  agentsToProto(info.Agents),
			ResponseDurationSeconds: info.ResponseDurationSeconds,
			Errors:                  info.Errors,
			FromCache:               info.FromCache,
			CacheSearchUuid:         info.CacheSearchID,
			CacheSearchCreatedAt:    info.CacheSearchCreatedAt,
		}
	}
	return result
}

func agentsToProto(agents map[int]v3.AgentDebugInfo) map[int64]*AgentDebugInfo {
	if agents == nil {
		return nil
	}
	result := make(map[int64]*AgentDebugInfo, len(agents))
	for id, info := range agents {
		result[int64(id)] = &AgentDebugInfo{
			Proposals:                proposalsMapToProto(info.Proposals),
			ProposalsCount:           int64(info.ProposalsCount),
			BadProposals:             convMap(info.BadProposals, same[string], convNum[int, int64]),
			FilteredProposals:        convMap(info.FilteredProposals, same[string], proposalsToProto),
			MergedFlightTermsSources: convMap(info.MergedFlightTermsSources, same[string], convNum[int, int64]),
		}
	}
	return result
}

func proposalsMapToProto(proposals map[v3.ProposalID]v3.ProposalDebugInfo) map[string]*ProposalDebugInfo {
	if proposals == nil {
		return nil
	}
	result := make(map[string]*ProposalDebugInfo, len(proposals))
	for id, info := range proposals {
		result[id] = &ProposalDebugInfo{
			AgencyPrice:  amountToProto(info.AgencyPrice),
			Multiplier:   info.Multiplier,
			Productivity: info.Productivity,
			FlightTerms:  flightTermsDebugInfoToProto(info.FlightTerms),
			Cashback:     cashbackDebugInfoToProto(info.Cashback),
		}
	}
	return result
}

func amountToProtoOpt(amount *currency.Amount) *Amount {
	if amount == nil {
		return nil
	}
	return amountToProto(*amount)
}

func amountToProto(amount currency.Amount) *Amount {
	return &Amount{
		CurrencyCode: Currency(Currency_value[amount.CurrencyCode.String()]),
		Value:        amount.Value,
	}
}

func flightTermsDebugInfoToProto(terms map[v3.FlightLegIndex]v3.FlightTermDebugInfo) map[int64]*FlightTermDebugInfo {
	if terms == nil {
		return nil
	}
	result := make(map[int64]*FlightTermDebugInfo, len(terms))
	for index, info := range terms {
		result[int64(index)] = &FlightTermDebugInfo{
			BaggageSource:      TermSource(info.BaggageSource),
			HandbagsSource:     TermSource(info.HandbagsSource),
			GateTechnicalStops: convArray(info.GateTechnicalStops, technicalStopToProto),
		}
	}
	return result
}

func technicalStopToProtoOpt(stop *delta.TechnicalStop) *TechnicalStop {
	if stop == nil {
		return nil
	}
	return technicalStopToProto(*stop)
}

func technicalStopToProto(stop delta.TechnicalStop) *TechnicalStop {
	return &TechnicalStop{
		AirportCode: string(stop.AirportCode),
	}
}

func cashbackDebugInfoToProto(info *v3.CashbackDebugInfo) *CashbackDebugInfo {
	if info == nil {
		return nil
	}
	return &CashbackDebugInfo{
		Amount:          amountToProtoOpt(info.Amount),
		LocalizedAmount: amountToProtoOpt(info.LocalizedAmount),
		Available:       info.Available,
	}
}

func proposalsToProto(proposals []v3.Proposal) *Proposals {
	if proposals == nil {
		return nil
	}
	result := make([]*Proposal, len(proposals))
	for i, proposal := range proposals {
		result[i] = &Proposal{
			Id:                proposal.ID,
			Price:             amountToProto(proposal.Price),
			PricePerPerson:    amountToProto(proposal.PricePerPerson),
			AgentId:           int64(proposal.AgentID),
			FlightTerms:       convMap(proposal.FlightTerms, convNum[int, int64], flightTermToProto),
			TransferTerms:     transferTermsToProto(proposal.TransferTerms),
			UnifiedPrice:      amountToProto(proposal.UnifiedPrice),
			Options:           proposalOptionsToProto(proposal.Options),
			Weight:            proposal.Weight,
			FromMainAirline:   proposal.FromMainAirline,
			Tags:              proposal.Tags,
			MinimumFare:       fateToProto(proposal.MinimumFare),
			IsWarmcache:       proposal.IsWarmcache,
			Cashback:          cashbackToProto(proposal.Cashback),
			CashbackPerPerson: cashbackToProto(proposal.CashbackPerPerson),
			AcceptedCards:     convArray(proposal.AcceptedCards, acceptedCardToProto),
		}
	}
	return &Proposals{
		Proposals: result,
	}
}

func flightTermToProto(term v3.FlightTerm) *FlightTerm {
	return &FlightTerm{
		FareCode:                   string(term.FareCode),
		TripClass:                  TripClass(term.TripClass),
		SeatsAvailable:             int32(term.SeatsAvailable),
		MarketingCarrierDesignator: flightDesignatorToProto(term.MarketingCarrierDesignator),
		Baggage:                    baggageToProto(term.Baggage),
		Handbags:                   baggageToProto(term.Handbags),
		AdditionalTariffInfo:       additionalTariffInfoToProto(term.AdditionalTariffInfo),
		IsCharter:                  term.IsCharter,
		Tags:                       term.Tags,
		MergedTermsInfo:            mergedTermsInfoToProto(term.MergedTermsInfo),
		MergedFromOtherProposals:   convMap(term.MergedFromOtherProposals, same[string], convNum[int, int64]),
	}
}

func flightDesignatorToProto(fd *base.FlightDesignator) *FlightDesignator {
	if fd == nil {
		return nil
	}
	return &FlightDesignator{
		Carrier:   string(fd.Carrier),
		AirlineId: string(fd.AirlineID),
		Number:    string(fd.Number),
	}
}

func additionalTariffInfoToProto(ti *v3.AdditionalTariffInfo) *AdditionalTariffInfo {
	if ti == nil {
		return nil
	}
	return &AdditionalTariffInfo{
		SeatAtPurchaseInfo:     tariffInfoToProto(ti.SeatAtPurchaseInfo),
		SeatAtRegistrationInfo: tariffInfoToProto(ti.SeatAtRegistrationInfo),
		ReturnBeforeFlight:     tariffInfoToProto(ti.ReturnBeforeFlight),
		ReturnAfterFlight:      tariffInfoToProto(ti.ReturnAfterFlight),
		ChangeBeforeFlight:     tariffInfoToProto(ti.ChangeBeforeFlight),
		ChangeAfterFlight:      tariffInfoToProto(ti.ChangeAfterFlight),
		FareName:               ti.FareName,
		Miles:                  ti.Miles,
	}
}

func mergedTermsInfoToProto(mti v3.MergedTermsInfo) *MergedTermsInfo {
	return &MergedTermsInfo{
		SeatAtRegistration: tariffMergeInfoToProto(mti.SeatAtRegistration),
		SeatAtPurchase:     tariffMergeInfoToProto(mti.SeatAtPurchase),
		ReturnBeforeFlight: tariffMergeInfoToProto(mti.ReturnBeforeFlight),
		ReturnAfterFlight:  tariffMergeInfoToProto(mti.ReturnAfterFlight),
		ChangeBeforeFlight: tariffMergeInfoToProto(mti.ChangeBeforeFlight),
		ChangeAfterFlight:  tariffMergeInfoToProto(mti.ChangeAfterFlight),
		Baggage:            baggageMergeInfoToProto(mti.Baggage),
		Handbags:           baggageMergeInfoToProto(mti.Handbags),
	}
}

func tariffMergeInfoToProto(info v3.TariffMergeInfo) *TariffMergeInfo {
	return &TariffMergeInfo{
		IsFromConfig: tariffMergeParamsToProto(info.IsFromConfig),
		Mismatch:     tariffMergeParamsToProto(info.Mismatch),
	}
}

func tariffMergeParamsToProto(tmp v3.TariffMergeParams) *TariffMergeParams {
	return &TariffMergeParams{
		Available:           tmp.Available,
		PenaltyCurrencyCode: tmp.PenaltyCurrencyCode,
		PenaltyValue:        tmp.PenaltyValue,
	}
}

func baggageMergeInfoToProto(info v3.BaggageMergeInfo) *BaggageMergeInfo {
	return &BaggageMergeInfo{
		IsFromConfig: baggageMergeParamsToProto(info.IsFromConfig),
		Mismatch:     baggageMergeParamsToProto(info.Mismatch),
	}
}

func baggageMergeParamsToProto(bmp v3.BaggageMergeParams) *BaggageMergeParams {
	return &BaggageMergeParams{
		Count:        bmp.Count,
		Weight:       bmp.Weight,
		TotalWeight:  bmp.TotalWeight,
		Height:       bmp.Height,
		Length:       bmp.Length,
		Width:        bmp.Width,
		SumDimension: bmp.SumDimension,
	}
}

func transferTermsToProto(terms [][]v3.TransferTerm) []*TransferTerms {
	return convArray(terms, func(v1 []v3.TransferTerm) *TransferTerms {
		if v1 == nil {
			return nil
		}
		return &TransferTerms{
			Terms: convArray(v1, func(v1 v3.TransferTerm) *TransferTerm {
				return &TransferTerm{
					IsVirtualInterline: v1.IsVirtualInterline,
					Tags:               v1.Tags,
				}
			}),
		}
	})
}

func proposalOptionsToProto(opts *v3.ProposalOptions) *ProposalOptions {
	if opts == nil {
		return nil
	}
	if opts.Hotel == nil {
		return &ProposalOptions{}
	}
	return &ProposalOptions{
		Hotel: &Hotel{
			Name:     opts.Hotel.Name,
			Stars:    opts.Hotel.Stars,
			RoomType: opts.Hotel.RoomType,
			Meals:    opts.Hotel.Meals,
		},
	}
}

func fateToProto(fare v3.Fare) *Fare {
	return &Fare{
		Code:               fare.Code,
		Baggage:            baggageToProto(fare.Baggage),
		Handbags:           baggageToProto(fare.Handbags),
		ReturnBeforeFlight: tariffInfoToProto(fare.ReturnBeforeFlight),
		ReturnAfterFlight:  tariffInfoToProto(fare.ReturnAfterFlight),
		ChangeBeforeFlight: tariffInfoToProto(fare.ChangeBeforeFlight),
		ChangeAfterFlight:  tariffInfoToProto(fare.ChangeAfterFlight),
		SeatAtPurchase:     tariffInfoToProto(fare.SeatAtPurchase),
		SeatAtRegistration: tariffInfoToProto(fare.SeatAtRegistration),
		FareName:           fare.FareName,
		Miles:              fare.Miles,
	}
}

func baggageToProto(baggage *base.Baggage) *Baggage {
	if baggage == nil {
		return nil
	}
	return &Baggage{
		Count:        int64(baggage.Count),
		Weight:       baggage.Weight,
		TotalWeight:  baggage.TotalWeight,
		Length:       baggage.Length,
		Width:        baggage.Width,
		Height:       baggage.Height,
		SumDimension: baggage.SumDimension,
	}
}

func tariffInfoToProto(info *v3.TariffInfo) *TariffInfo {
	if info == nil {
		return nil
	}
	return &TariffInfo{
		Available:    info.Available,
		Penalty:      amountToProtoOpt(info.Penalty),
		IsFromConfig: info.IsFromConfig,
	}
}

func cashbackToProto(cashback *v3.Cashback) *Cashback {
	if cashback == nil {
		return nil
	}
	return &Cashback{
		LocalizedAmount: amountToProtoOpt(cashback.LocalizedAmount),
		Available:       cashback.Available,
	}
}

func acceptedCardToProto(card v3.AcceptedCard) *AcceptedCard {
	return &AcceptedCard{
		Region: card.Region,
		System: card.System,
	}
}

func ticketToProto(ticket v3.Ticket) *Ticket {
	var proposals []*Proposal
	proposalsObj := proposalsToProto(ticket.Proposals)
	if proposalsObj != nil {
		proposals = proposalsObj.Proposals
	}
	return &Ticket{
		Segments:   convArray(ticket.Segments, segmentToProto),
		Proposals:  proposals,
		Signature:  ticket.Signature,
		Popularity: ticket.Popularity,
		Score:      ticket.Score,
		Hashsum:    ticket.HashSum,
		Tags:       ticket.Tags,
		Badges:     convArray(ticket.Badges, badgeInfoToProto),
		ExtraFares: convMap(ticket.ExtraFares, same[string], fareProposalsToProto),
		FilteredBy: ticket.FilteredBy,
	}
}

func ticketToProtoOpt(ticket *v3.Ticket) *Ticket {
	if ticket == nil {
		return nil
	}
	return ticketToProto(*ticket)
}

func segmentToProto(segment v3.Segment) *Segment {
	return &Segment{
		Flights:   convArray(segment.FlightLegs, convNum[int, int64]),
		Transfers: convArray(segment.Transfers, transferToProto),
		Tags:      segment.Tags,
	}
}

func transferToProto(transfer v3.Transfer) *Transfer {
	return &Transfer{
		VisaRules: &VisaRules{
			Required: transfer.VisaRules.Required,
		},
		RecheckBaggage: transfer.RecheckBaggage,
		NightTransfer:  transfer.NightTransfer,
		Tags:           transfer.Tags,
	}
}

func badgeInfoToProto(info v3.BadgeInfo) *BadgeInfo {
	return &BadgeInfo{
		Type:   info.Type,
		Scores: info.Scores,
		Meta: &BadgeInfoMeta{
			Name:     convMap(info.Meta.Name, convString[base.LanguageCode], same[string]),
			Priority: int64(info.Meta.Priority),
			Position: int64(info.Meta.Position),
			Limit:    int64(info.Meta.Limit),
			Colors: &Colors{
				Light: info.Meta.Colors.Light,
				Dark:  info.Meta.Colors.Dark,
			},
		},
	}
}

func fareProposalsToProto(proposals []v3.FareProposal) *FareProposals {
	if proposals == nil {
		return nil
	}
	return &FareProposals{
		Proposals: convArray(proposals, func(v1 v3.FareProposal) *FareProposal {
			return &FareProposal{
				ProposalId: v1.ID,
				Index:      int64(v1.Index),
			}
		}),
	}
}

func softResponseToProto(response *v3.SoftResponse) *SoftResponse {
	if response == nil {
		return nil
	}
	return &SoftResponse{
		FiltersApplied: response.FiltersApplied,
		Tickets:        convArray(response.Tickets, ticketToProto),
	}
}

func directFlightsToProto(df v3.DirectFlights) *DirectFlights {
	return &DirectFlights{
		Carrier:        df.Carrier,
		Carriers:       df.Carriers,
		CheapestTicket: ticketToProto(df.CheapestTicket),
		Schedule: convArray(df.Schedule, func(v1 []v3.Schedule) *ScheduleList {
			if v1 == nil {
				return nil
			}
			return &ScheduleList{
				List: convArray(v1, func(v1 v3.Schedule) *Schedule {
					return &Schedule{
						Time:              v1.Time,
						Datetime:          v1.DateTime,
						TicketsSignatures: v1.TicketsSignatures,
					}
				}),
			}
		}),
	}
}

func flightLegToProto(leg v3.FlightLeg) *FlightLeg {
	return &FlightLeg{
		Origin:                     string(leg.Origin),
		Destination:                string(leg.Destination),
		LocalDepartureDateTime:     leg.LocalDepartureDateTime.String(),
		LocalArrivalDateTime:       leg.LocalArrivalDateTime.String(),
		DepartureUnixTimestamp:     leg.DepartureUnixTimestamp,
		ArrivalUnixTimestamp:       leg.ArrivalUnixTimestamp,
		OperatingCarrierDesignator: flightDesignatorToProto(&leg.OperatingCarrierDesignator),
		Equipment:                  baseEquipmentToProto(leg.Equipment),
		TechnicalStops:             convArray(leg.TechnicalStops, technicalStopToProtoOpt),
		Signature:                  leg.Signature,
		Tags:                       leg.Tags,
	}
}

func equipmentToProto(equipment v3.Equipment) *Equipment {
	return &Equipment{
		Code: equipment.Code,
		Type: EquipmentType(equipment.Type),
		Name: equipment.Name,
	}
}

func baseEquipmentToProto(equipment base.Equipment) *Equipment {
	return &Equipment{
		Code: string(equipment.Code),
		Type: EquipmentType(equipment.Type),
		Name: equipment.Name,
	}
}

func airlineInfoToProto(info v3.AirlineInfo) *AirlineInfo {
	return &AirlineInfo{
		Iata:       string(info.IATA),
		IsLowcost:  info.IsLowcost,
		Name:       localizableContextStringToProto(info.Name),
		AllianceId: int64(info.AllianceID),
		SiteName:   info.SiteName,
		BrandColor: info.BrandColor,
	}
}

func localizableContextStringToProto(name base.LocalizableContextString) map[string]*MapStringString {
	if name == nil {
		return nil
	}

	result := make(map[string]*MapStringString, len(name))
	for k1, v1 := range name {
		if v1 != nil {
			result[k1.String()] = &MapStringString{
				Map: v1,
			}
		}
	}
	return result
}

func placesToProto(places base.Places) *Places {
	return &Places{
		Airports: convMap(places.Airports, convString[iata.LocationIATACode],
			func(v1 base.AirportInfo) *AirportInfo {
				return &AirportInfo{
					Name:          localizableContextStringToProto(v1.Name),
					Code:          string(v1.Code),
					CityCode:      string(v1.CityCode),
					MetroAreaCode: string(v1.MetroAreaCode),
					Coordinates: &GeoPoint{
						Lat: v1.Coordinates.Lat,
						Lng: v1.Coordinates.Lng,
					},
					HasTransitZone:      pointerBoolToProto(v1.HasTransitZone),
					TransitWorkHoursMin: int64(v1.TransitWorkHoursMin),
					TransitWorkHoursMax: int64(v1.TransitWorkHoursMax),
				}
			}),
		Cities: convMap(places.Cities, convString[iata.LocationIATACode],
			func(v1 base.CityInfo) *CityInfo {
				return &CityInfo{
					Code:     string(v1.Code),
					Name:     localizableContextStringToProto(v1.Name),
					Country:  string(v1.Country),
					Timezone: v1.Timezone,
					Airports: convArray(v1.Airports, convString[iata.LocationIATACode]),
				}
			}),
		Countries: convMap(places.Countries, convString[iata.CountryCode],
			func(v1 base.CountryInfo) *CountryInfo {
				return &CountryInfo{
					Code:        string(v1.Code),
					Name:        localizableContextStringToProto(v1.Name),
					UnifiedVisa: v1.UnifiedVisa,
				}
			}),
		MetroAreas: convMap(places.MetroAreas, convString[iata.LocationIATACode],
			func(v1 base.MetroAreaInfo) *MetroAreaInfo {
				return &MetroAreaInfo{
					Code:     string(v1.Code),
					Airports: convArray(v1.Airports, convString[iata.LocationIATACode]),
					Timezone: v1.Timezone,
				}
			}),
		AirportsToMetro: convMap(places.AirportsToMetro, convString[iata.LocationIATACode], convString[iata.LocationIATACode]),
	}
}

func pointerBoolToProto(bool base.PointerBool) *OptBool {
	return &OptBool{
		Value:     bool.IsTrue(),
		IsUnknown: bool.IsUnknown(),
	}
}

func agentInfoToProto(info v3.AgentInfo) *AgentInfo {
	return &AgentInfo{
		Id:             int64(info.ID),
		GateName:       info.GateName,
		Label:          localizableContextStringToProto(info.Label),
		PaymentMethods: info.PaymentMethods,
		MobileVersion:  info.MobileVersion,
		HideProposals:  info.HideProposals,
		Assisted:       info.Assisted,
		MobileType:     info.MobileType,
		AirlineIatas:   info.AirlineIATAs,
	}
}

func allianceToProto(alliance v3.Alliance) *Alliance {
	return &Alliance{
		Id:   int64(alliance.ID),
		Name: alliance.Name,
	}
}

func degradedBoundariesToProto(bound *boundaries.DegradedBoundaries) *DegradedBoundaries {
	if bound == nil {
		return nil
	}

	var airports map[int64]*DegradedAirportsBoundaries
	if bound.Airports != nil {
		airports = make(map[int64]*DegradedAirportsBoundaries, len(bound.Airports))
		for i, airportsBoundaries := range bound.Airports {
			airports[int64(i)] = &DegradedAirportsBoundaries{
				Arrival:   convMap(airportsBoundaries.Arrival, convString[iata.LocationIATACode], filterPriceToProto),
				Departure: convMap(airportsBoundaries.Departure, convString[iata.LocationIATACode], filterPriceToProto),
			}
		}
	}

	var baggage *FilterBaggageBoundaries
	if bound.Baggage != nil {
		baggage = &FilterBaggageBoundaries{
			FullBaggage:  filterPriceToProto(bound.Baggage.FullBaggage),
			NoBaggage:    filterPriceToProto(bound.Baggage.NoBaggage),
			LargeHandbag: filterPriceToProto(bound.Baggage.LargeHandbag),
		}
	}

	var departureArrivalTime map[int64]*DegradedTimeBoundaries
	if bound.DepartureArrivalTime != nil {
		departureArrivalTime = make(map[int64]*DegradedTimeBoundaries, len(bound.DepartureArrivalTime))
		for i, timeBoundaries := range bound.DepartureArrivalTime {
			departureArrivalTime[int64(i)] = &DegradedTimeBoundaries{
				ArrivalDate:   convMap(timeBoundaries.ArrivalDate, convStringer[datetime.Date], filterPriceToProto),
				ArrivalTime:   dateTimeRangeBoundariesToProtoOpt(timeBoundaries.ArrivalTime),
				DepartureTime: dateTimeRangeBoundariesToProtoOpt(timeBoundaries.DepartureTime),
				TripDuration:  rangeBoundariesToProtoOpt((*times.RangeBoundaries)(timeBoundaries.TripDuration)),
			}
		}
	}

	var returnTicket *DegradedReturnTicketBoundaries
	if bound.ReturnTicket != nil {
		returnTicket = &DegradedReturnTicketBoundaries{
			Available: filterPriceToProto(bound.ReturnTicket.Available),
			Free:      filterPriceToProto(bound.ReturnTicket.Free),
		}
	}

	var changeTicket *DegradedReturnTicketBoundaries
	if bound.ChangeTicket != nil {
		changeTicket = &DegradedReturnTicketBoundaries{
			Available: filterPriceToProto(bound.ChangeTicket.Available),
			Free:      filterPriceToProto(bound.ChangeTicket.Free),
		}
	}

	return &DegradedBoundaries{
		Agents:                           convMap(bound.Agents, convNum[int, int64], filterPriceToProto),
		Airlines:                         convMap(bound.Airlines, convStringer[iata.AirlineID], filterPriceToProto),
		Alliances:                        convMap(bound.Alliances, convNum[int, int64], filterPriceToProto),
		HasInterlines:                    filterBoolToProto(bound.HasInterlines),
		HasLowcosts:                      filterBoolToProto(bound.HasLowcosts),
		Airports:                         airports,
		SameDepartureArrivalAirport:      convMap(bound.SameDepartureArrivalAirport, convString[iata.LocationIATACode], filterPriceToProto),
		Baggage:                          baggage,
		Equipments:                       convMap(bound.Equipments, same[string], filterPriceToProto),
		PaymentMethods:                   convMap(bound.PaymentMethods, same[string], filterPriceToProto),
		Price:                            floatRangeToProtoOpt((*filter.FloatRange)(bound.Price)),
		DepartureArrivalTime:             departureArrivalTime,
		ReturnTicket:                     returnTicket,
		ChangeTicket:                     changeTicket,
		TransfersCount:                   convMap(bound.TransfersCount, same[int64], filterPriceToProto),
		TransfersDuration:                transferDurationBoundariesToProto(bound.TransfersDuration),
		TransfersAirports:                convMap(bound.TransfersAirports, convStringer[iata.LocationIATACode], filterPriceToProto),
		TransfersCountries:               convMap(bound.TransfersCountries, same[string], filterPriceToProto),
		HasTransfersWithAirportChange:    filterBoolToProto(bound.HasTransfersWithAirportChange),
		HasTransfersWithBaggageRecheck:   filterBoolToProto(bound.HasTransfersWithBaggageRecheck),
		HasTransfersWithVisa:             filterBoolToProto(bound.HasTransfersWithVisa),
		HasTransfersWithVirtualInterline: filterBoolToProto(bound.HasTransfersWithVirtualInterline),
		HasCovidRestrictions:             filterBoolToProto(bound.HasCovidRestrictions),
		HasNightTransfers:                filterBoolToProto(bound.HasNightTransfers),
		HasConvenientTransfers:           filterBoolToProto(bound.HasConvenientTransfers),
		HasShortLayoverTransfers:         filterBoolToProto(bound.HasShortLayoverTransfers),
		HasLongLayoverTransfers:          filterBoolToProto(bound.HasLongLayoverTransfers),
	}
}

func filterPriceToProto(price *filter.Price) *FilterPrice {
	if price == nil {
		return nil
	}
	return &FilterPrice{
		EnableMinPrice:  price.EnableMinPrice,
		DisableMinPrice: price.DisableMinPrice,
	}
}

func filterBoolToProto(bool *filter.Bool) *FilterBool {
	if bool == nil {
		return nil
	}
	return &FilterBool{
		EnableMinPrice:  bool.EnableMinPrice,
		DisableMinPrice: bool.DisableMinPrice,
	}
}

func floatRangeToProtoOpt(fr *filter.FloatRange) *PriceBoundaries {
	if fr == nil {
		return nil
	}
	return floatRangeToProto(*fr)
}

func floatRangeToProto(fr filter.FloatRange) *PriceBoundaries {
	return &PriceBoundaries{
		Min: fr.Min,
		Max: fr.Max,
	}
}

func dateTimeRangeBoundariesToProto(dtrb times.DateTimeRangeBoundaries) *DateTimeRangeBoundaries {
	return &DateTimeRangeBoundaries{
		Min:         dtrb.Min.String(),
		Max:         dtrb.Max.String(),
		Buckets:     convMap(dtrb.Buckets, convStringer[datetime.DateTime], same[float64]),
		BucketWidth: dtrb.BucketWidth,
	}
}

func dateTimeRangeBoundariesToProtoOpt(dtrb *times.DateTimeRangeBoundaries) *DateTimeRangeBoundaries {
	if dtrb == nil {
		return nil
	}
	return dateTimeRangeBoundariesToProto(*dtrb)
}

func rangeBoundariesToProtoOpt(rb *times.RangeBoundaries) *RangeBoundaries {
	if rb == nil {
		return nil
	}
	return rangeBoundariesToProto(*rb)
}

func rangeBoundariesToProto(rb times.RangeBoundaries) *RangeBoundaries {
	return &RangeBoundaries{
		Min:         rb.Min,
		Max:         rb.Max,
		Buckets:     rb.Buckets,
		BucketWidth: rb.BucketWidth,
	}
}

func transferDurationBoundariesToProto(tdb *transfers.TransferDurationBoundaries) *TransferDurationBoundaries {
	if tdb == nil {
		return nil
	}
	return &TransferDurationBoundaries{
		Min: tdb.Min,
		Max: tdb.Max,
	}
}

func boundariesToProto(bound *boundaries.Boundaries) *Boundaries {
	if bound == nil {
		return nil
	}

	var airports map[int64]*AirportsBoundaries
	if bound.Airports != nil {
		airports = make(map[int64]*AirportsBoundaries, len(bound.Airports))
		for i, airportsBoundaries := range bound.Airports {
			airports[int64(i)] = &AirportsBoundaries{
				Arrival:   convMap(airportsBoundaries.Arrival, convString[iata.LocationIATACode], same[float64]),
				Departure: convMap(airportsBoundaries.Departure, convString[iata.LocationIATACode], same[float64]),
			}
		}
	}

	var departureArrivalTime map[int64]*TimeBoundaries
	if bound.DepartureArrivalTime != nil {
		departureArrivalTime = make(map[int64]*TimeBoundaries, len(bound.DepartureArrivalTime))
		for i, timeBoundaries := range bound.DepartureArrivalTime {
			departureArrivalTime[int64(i)] = &TimeBoundaries{
				ArrivalDate:   convMap(timeBoundaries.ArrivalDate, convStringer[datetime.Date], same[float64]),
				ArrivalTime:   dateTimeRangeBoundariesToProto(timeBoundaries.ArrivalTime),
				DepartureTime: dateTimeRangeBoundariesToProto(timeBoundaries.DepartureTime),
				TripDuration:  rangeBoundariesToProto((times.RangeBoundaries)(timeBoundaries.TripDuration)),
			}
		}
	}

	return &Boundaries{
		Agents:                      convMap(bound.Agents, convNum[int, int64], same[float64]),
		Airlines:                    convMap(bound.Airlines, convStringer[iata.AirlineID], same[float64]),
		Alliances:                   convMap(bound.Alliances, convNum[int, int64], same[float64]),
		HasInterlines:               bound.HasInterlines,
		HasLowcosts:                 bound.HasLowcosts,
		Airports:                    airports,
		SameDepartureArrivalAirport: convMap(bound.SameDepartureArrivalAirport, convStringer[iata.LocationIATACode], same[float64]),
		Baggage: &BaggageBoundaries{
			FullBaggage:  bound.Baggage.FullBaggage,
			NoBaggage:    bound.Baggage.NoBaggage,
			LargeHandbag: bound.Baggage.LargeHandbag,
		},
		Equipments:           bound.Equipments,
		PaymentMethods:       bound.PaymentMethods,
		Price:                floatRangeToProto(filter.FloatRange(bound.Price)),
		DepartureArrivalTime: departureArrivalTime,
		ReturnTicket: &ReturnBoundaries{
			Available: bound.ReturnTicket.Available,
			Free:      bound.ReturnTicket.Free,
		},
		ChangeTicket: &ChangeBoundaries{
			Available: bound.ChangeTicket.Available,
			Free:      bound.ChangeTicket.Free,
		},
		TransfersCount:                   bound.TransfersCount,
		TransfersDuration:                transferDurationBoundariesToProto(bound.TransfersDuration),
		TransfersAirports:                convMap(bound.TransfersAirports, convStringer[iata.LocationIATACode], same[float64]),
		TransfersCountries:               bound.TransfersCountries,
		HasTransfersWithAirportChange:    bound.HasTransfersWithAirportChange,
		HasTransfersWithBaggageRecheck:   bound.HasTransfersWithBaggageRecheck,
		HasTransfersWithVisa:             bound.HasTransfersWithVisa,
		HasTransfersWithVirtualInterline: bound.HasTransfersWithVirtualInterline,
		HasCovidRestrictions:             bound.HasCovidRestrictions,
		HasNightTransfers:                bound.HasNightTransfers,
		HasConvenientTransfers:           bound.HasConvenientTransfers,
		HasShortLayoverTransfers:         bound.HasShortLayoverTransfers,
		HasLongLayoverTransfers:          bound.HasLongLayoverTransfers,
	}
}

func filterState(state *filter.State) *FilterState {
	if state == nil {
		return nil
	}
	return &FilterState{
		Agents:                          convArray(state.Agents, convNum[int, int64]),
		Airlines:                        state.Airlines,
		Alliances:                       convArray(state.Alliances, convNum[int, int64]),
		WithoutInterlines:               state.WithoutInterlines,
		WithoutLowcosts:                 state.WithoutLowcosts,
		Segments:                        convMap(state.Segments, convNum[int, int64], segmentFilterToProto),
		WithSameDepartureArrivalAirport: state.WithSameDepartureArrivalAirport,
		Equipments:                      state.Equipments,
		PaymentMethods:                  state.PaymentMethods,
		PinFlightSignatures:             state.PinFlightSignatures,
		Price: convArray(state.Price, func(v1 filter.FloatRange) *FloatRange {
			return (*FloatRange)(floatRangeToProto(v1))
		}),
		TransfersCount:                   convArray(state.TransfersCount, convNum[int, int64]),
		TransfersDuration:                convArray(state.TransfersDuration, rangeToProto),
		TransfersWithoutAirportChange:    state.TransfersWithoutAirportChange,
		TransfersWithoutBaggageRecheck:   state.TransfersWithoutBaggageRecheck,
		TransfersWithoutVisa:             state.TransfersWithoutVisa,
		TransfersWithoutVirtualInterline: state.TransfersWithoutVirtualInterline,
		ConvenientTransfers:              state.ConvenientTransfers,
		WithoutNightTransfers:            state.WithoutNightTransfers,
		WithoutShortLayover:              state.WithoutShortLayover,
		WithoutLongLayover:               state.WithoutLongLayover,
		TransfersAirports:                state.TransfersAirports,
		TransfersCountries:               state.TransfersCountries,
		WithoutCovidRestrictions:         state.WithoutCovidRestrictions,
		Baggage:                          state.Baggage,
		TimeBuckets:                      timeBucketsToProto(state.TimeBuckets),
		ReturnBeforeFlight:               state.ReturnBeforeFlight,
		ChangeBeforeFlight:               state.ChangeBeforeFlight,
	}
}

func segmentFilterToProto(f filter.SegmentFilter) *SegmentFilter {
	return &SegmentFilter{
		AirportsArrival:   f.AirportsArrival,
		AirportsDeparture: f.AirportsDeparture,
		ArrivalTime:       convArray(f.ArrivalTime, dateTimeOrTimeRangeToProto),
		ArrivalDate:       convArray(f.ArrivalDate, convStringer[datetime.Date]),
		DepartureTime:     convArray(f.DepartureTime, dateTimeRangeToProto),
		TripDuration:      convArray(f.TripDuration, rangeToProto),
	}
}

func dateTimeOrTimeRangeToProto(d filter.DateTimeOrTimeRange) *DateTimeOrTimeRange {
	return &DateTimeOrTimeRange{
		Min: d.Min,
		Max: d.Max,
	}
}

func dateTimeRangeToProto(d filter.DateTimeRange) *DateTimeRange {
	var min, max int64
	if d.Min != nil {
		min = d.Min.Unix()
	}
	if d.Max != nil {
		max = d.Max.Unix()
	}

	return &DateTimeRange{
		Min: min,
		Max: max,
	}
}

func rangeToProto(r filter.Range) *Range {
	return &Range{
		Min: r.Min,
		Max: r.Max,
	}
}

func timeBucketsToProto(buckets *filter.TimeBuckets) *TimeBuckets {
	if buckets == nil {
		return nil
	}
	return &TimeBuckets{
		ArrivalTimeBucketWidth:      int64(buckets.ArrivalTimeBucketWidth),
		DepartureTimeBucketWidth:    int64(buckets.DepartureTimeBucketWidth),
		TripDurationTimeBucketWidth: int64(buckets.TripDurationTimeBucketWidth),
	}
}

func convMap[K1 comparable, K2 comparable, V1 any, V2 any](from map[K1]V1, k func(K1) K2, v func(V1) V2) map[K2]V2 {
	if from == nil {
		return nil
	}
	result := make(map[K2]V2, len(from))
	for k1, v1 := range from {
		result[k(k1)] = v(v1)
	}
	return result
}

func convArray[V1 any, V2 any](from []V1, conv func(V1) V2) []V2 {
	if from == nil {
		return nil
	}
	result := make([]V2, len(from))
	for i, v1 := range from {
		result[i] = conv(v1)
	}
	return result
}

func same[A any](val A) A {
	return val
}

func convString[V1 ~string](val V1) string {
	return string(val)
}

func convStringer[V1 fmt.Stringer](val V1) string {
	return val.String()
}

type Number interface {
	constraints.Integer | constraints.Float
}

func convNum[A Number, B Number](a A) B {
	return B(a)
}
