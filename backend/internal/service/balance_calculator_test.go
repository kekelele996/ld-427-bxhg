package service

import "testing"

func TestCalculateAvailable(t *testing.T) {
	tests := []struct {
		name   string
		total  float64
		spent  float64
		frozen float64
		want   float64
	}{
		{name: "fresh budget", total: 1000, spent: 0, frozen: 0, want: 1000},
		{name: "with spent and frozen", total: 1000, spent: 300, frozen: 200, want: 500},
		{name: "negative guard", total: 500, spent: 300, frozen: 300, want: -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CalculateAvailable(tt.total, tt.spent, tt.frozen); got != tt.want {
				t.Fatalf("CalculateAvailable = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateVariance(t *testing.T) {
	tests := []struct {
		spent  float64
		budget float64
		want   float64
	}{
		{spent: 0, budget: 100, want: -100},
		{spent: 80, budget: 100, want: -20},
		{spent: 120, budget: 100, want: 20},
	}
	for _, tt := range tests {
		if got := CalculateVariance(tt.spent, tt.budget); got != tt.want {
			t.Fatalf("CalculateVariance(%v,%v) = %v, want %v", tt.spent, tt.budget, got, tt.want)
		}
	}
}

func TestCalculateUnpaid(t *testing.T) {
	tests := []struct {
		payable float64
		paid    float64
		want    float64
	}{
		{payable: 1000, paid: 300, want: 700},
		{payable: 1000, paid: 1000, want: 0},
		{payable: 1000, paid: 0, want: 1000},
	}
	for _, tt := range tests {
		if got := CalculateUnpaid(tt.payable, tt.paid); got != tt.want {
			t.Fatalf("CalculateUnpaid(%v,%v) = %v, want %v", tt.payable, tt.paid, got, tt.want)
		}
	}
}
