using System;
using System.Text.Json;
using System.Text.Json.Serialization;

namespace BlazorApp.Services;

public class MeasurementService
{
	private readonly HttpClient httpClient;
	public Collection[] collections = Array.Empty<Collection>();

	public MeasurementService(HttpClient _httpClient)
	{
		httpClient = _httpClient;
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

	// <summary>
	// Creates a new collection of measurements.
	// </summary>
	public async Task<(bool, string)> CreateMeasurementCollectionAsync(string name, string? description = "")
	{
		var content = new
		{
			name,
			description,
		};
		try
		{
			HttpResponseMessage response = await httpClient.PostAsJsonAsync("collections", content);

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to create measurement collection: {response.StatusCode}");
				return (false, response.Content.ReadAsStringAsync().Result);
			}
			return (true, "");
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error creating measurement collection: {ex.Message}");
			return (false, ex.Message);
		}
	}

	// <summary>
	// Fetches all collections
	// </summary>
	public async Task<Collection[]> FetchCollectionsAsync()
	{
		try
		{
			HttpResponseMessage response = await httpClient.GetAsync("collections");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch collections: {response.StatusCode}");
				return Array.Empty<Collection>();
			}

			string jsonString = await response.Content.ReadAsStringAsync();

			collections = JsonSerializer.Deserialize<Collection[]>(jsonString) ?? Array.Empty<Collection>();
			return collections;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching collections: {ex.Message}");
			return Array.Empty<Collection>();
		}
	}

	// <summary>
	// Deletes a collection by its ID.
	// </summary>
	public async Task<bool> DeleteCollectionAsync(int id)
	{
		try
		{
			HttpResponseMessage response = await httpClient.DeleteAsync($"collections?collection_id={id}");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to delete collection: {response.StatusCode}");
			}
			// Delete the collection from the local array
			collections = collections.Where(c => c.Id != id).ToArray();
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error deleting collection: {ex.Message}");
		}
		return false;
	}

	public int[] GetCollectionIds()
	{
		return collections.Select(c => c.Id).ToArray();
	}

	public Collection GetCollectionById(int id)
	{
		return collections.FirstOrDefault(c => c.Id == id) ?? new Collection();
	}
}

public class Collection
{
	[JsonPropertyName("id")]
	public int Id { get; set; }

	[JsonPropertyName("name")]
	public string Name { get; set; } = "";

	[JsonPropertyName("description")]
	public string Description { get; set; } = "";
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