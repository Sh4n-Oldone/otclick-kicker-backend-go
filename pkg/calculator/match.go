package calculator

import (
	"context"
	"math"
)

func MatchRaitingCalculation(ctx context.Context, score1, score2, rating11, rating21, rating12, rating22 int) (int, int, int, int, error){
	// average rate for team1
	averRating1 := rating11
	if rating12 > 0 {
		if rating11 > rating12 {
			averRating1 = (2 * rating11 + rating12) / 3  
		} else {
			averRating1 = (2 * rating12 + rating11) / 3
		}
	}

	// average rate for team2
	averRating2 := rating21
	if rating22 > 0 {
		if rating21 > rating22 {
			averRating2 = (2 * rating21 + rating22) / 3  
		} else {
			averRating2 = (2 * rating22 + rating21) / 3
		}
	}

	// rating diff basis depends on match result
	var d int
	if score1 > score2 {
		d = averRating2 - averRating1
	} else {
		d = averRating1 - averRating2
	}
	// calculate diff
	// rd := math.Abs(float64(averRating1) - float64(averRating2))
	rd := float64(d / 400)
	rd = math.Pow(10, rd)
	rd = rd + 1
	rd = 1 / rd
	rd = 1 - rd
	rd = rd * 6
	rd = rd * math.Abs(float64(score1) - float64(score2))
	rd = math.Round(rd)
	x := int(rd)

	// ratings recalculate
	if (score1 > score2) {
		rating11 += x
		rating21 -= x
		if (rating12 > 0 && rating22 > 0) {
			rating12 += x
			rating22 -= x
		}
	} else {
		rating11 -= x
		rating21 += x
		if (rating12 > 0 && rating22 > 0) {
			rating12 -= x
			rating22 += x
		}
	}			
	
	return rating11, rating21, rating12, rating22, nil
}
