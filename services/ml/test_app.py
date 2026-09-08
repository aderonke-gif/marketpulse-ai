import unittest

from app import analyze


class AnalyzeTest(unittest.TestCase):
    def test_positive_earnings_event(self):
        result = analyze({
            "id": "n-1", "ticker": "LLY",
            "headline": "Lilly beats Q1 earnings estimates and raises guidance",
            "body": "Revenue growth was strong after record demand.",
        })
        self.assertEqual(result["sentiment"], "positive")
        self.assertIn("earnings", result["events"])
        self.assertIn("guidance_change", result["events"])
        self.assertGreater(result["confidence"], 0.7)

    def test_negative_regulatory_event(self):
        result = analyze({
            "id": "n-2", "ticker": "XYZ",
            "headline": "SEC investigation follows weak quarter",
            "body": "The company reported a loss and cut its outlook.",
        })
        self.assertEqual(result["sentiment"], "negative")
        self.assertIn("regulatory", result["events"])

    def test_neutral_copy(self):
        result = analyze({
            "headline": "Company schedules investor presentation",
            "body": "Management will speak at an industry conference Tuesday.",
        })
        self.assertEqual(result["sentiment"], "neutral")


if __name__ == "__main__":
    unittest.main()

