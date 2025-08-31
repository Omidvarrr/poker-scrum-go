package services

import (
	"awesomeProject1/internal/dto"
	"awesomeProject1/internal/repositories"
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

type GreetingService struct {
	quotes       []string
	lastGreeting map[int]greetingCache
	userRepo     repositories.UserRepository
}

type greetingCache struct {
	message   string
	quote     string
	timestamp time.Time
}

func NewGreetingService(userRepo repositories.UserRepository) *GreetingService {
	service := &GreetingService{
		userRepo:     userRepo,
		lastGreeting: make(map[int]greetingCache),
	}
	service.loadQuotes()
	return service
}

func (gs *GreetingService) GetGreeting(userID int) (dto.GreetingResponse, error) {
	now := time.Now()

	user, err := gs.userRepo.GetUserById(userID)
	if err != nil {
		return dto.GreetingResponse{}, err
	}
	userName := user.FirstName

	if cached, exists := gs.lastGreeting[userID]; exists {
		if now.Sub(cached.timestamp) < 15*time.Minute {
			return dto.GreetingResponse{
				Message: cached.message,
				Quote:   cached.quote,
			}, nil
		}
	}

	message := gs.generateGreeting(userName, now)
	quote := gs.getRandomQuote()

	gs.lastGreeting[userID] = greetingCache{
		message:   message,
		quote:     quote,
		timestamp: now,
	}

	return dto.GreetingResponse{
		Message: message,
		Quote:   quote,
	}, nil
}

func (gs *GreetingService) generateGreeting(userName string, now time.Time) string {
	hour := now.Hour()
	var greetingOptions []string

	if hour >= 5 && hour < 12 {
		greetingOptions = []string{
			"Morning, %s! Ready to plan?",
			"Hey %s! Fresh start?",
			"Hello, %s! What's brewing?",
			"Rise and shine, %s!",
			"Morning, %s! Let's roll!",
			"Hey %s! Coffee and code?",
			"Hello, %s! New day ahead!",
		}
	} else if hour >= 12 && hour < 17 {
		greetingOptions = []string{
			"Hey %s! Afternoon vibes!",
			"Hello, %s! Still crushing it?",
			"Afternoon, %s! What's next?",
			"Hey %s! Midday momentum?",
			"Hi %s! Keep pushing!",
			"Hello, %s! In the zone?",
			"Hey %s! Let's build!",
		}
	} else if hour >= 17 && hour < 22 {
		greetingOptions = []string{
			"Evening, %s! Still at it?",
			"Hey %s! Wrapping up?",
			"Hello, %s! Night shift?",
			"Evening, %s! What's cooking?",
			"Hey %s! Second wind?",
			"Hi %s! Evening plans?",
			"Hello, %s! Still going strong?",
		}
	} else {
		greetingOptions = []string{
			"Night owl, %s!",
			"Hey %s! Burning midnight oil?",
			"Late night, %s! Still up?",
			"Hello, %s! Night coding?",
			"Hey %s! Can't sleep?",
			"Hi %s! Quiet hours!",
			"Hello, %s! Stars are out!",
		}
	}

	selectedGreeting := greetingOptions[rand.Intn(len(greetingOptions))]
	return fmt.Sprintf(selectedGreeting, userName)
}

func (gs *GreetingService) getRandomQuote() string {
	if len(gs.quotes) == 0 {
		return "Code is poetry written in the language of logic"
	}
	return gs.quotes[rand.Intn(len(gs.quotes))]
}

func (gs *GreetingService) loadQuotes() {
	file, err := os.Open("qoutes.md")
	if err != nil {
		gs.quotes = []string{"Code is poetry written in the language of logic"}
		return
	}
	defer file.Close()

	var quotes []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			quotes = append(quotes, line)
		}
	}

	if len(quotes) == 0 {
		gs.quotes = []string{"Code is poetry written in the language of logic"}
	} else {
		gs.quotes = quotes
	}
}
