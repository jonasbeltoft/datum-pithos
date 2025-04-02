using System;
using System.Text.Json;
using Microsoft.AspNetCore.Mvc;

namespace BlazorApp.Services;

public class MeasurementService
{
	private readonly HttpClient httpClient;

	public MeasurementService(HttpClient _httpClient)
	{
		httpClient = _httpClient;
		Console.WriteLine("hello controller");
	}

	// <summary>
	// Fetches measurement data from the API.
	// </summary>
	public async Task<MeasurementCollection?> FetchMeasurementsAsync()
	{
		try
		{
			HttpResponseMessage response = await httpClient.GetAsync("measurements");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch data: {response.StatusCode}");
				return null;
			}

			string jsonString = await response.Content.ReadAsStringAsync();
			return JsonSerializer.Deserialize<MeasurementCollection>(jsonString, new JsonSerializerOptions { PropertyNameCaseInsensitive = true });
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching measurements: {ex.Message}");
			return null;
		}
	}
}

public class MeasurementCollection
{
	public MeasurementHeader[] headers = [];

	public Measurement[] entries = [];
}

public class MeasurementHeader
{
	public string name = "";

	public string unit = "";
}
public class Measurement
{
	public string[] measurements = [];
}