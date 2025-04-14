using System;
using System.Text;
using System.Text.Json;
using System.Text.Json.Serialization;
using BlazorApp.Components.Pages.Measurements;

namespace BlazorApp.Services;

public class MeasurementService
{
	private readonly HttpClient httpClient;
	public Collection[] collections = Array.Empty<Collection>();
	public Unit[] units = Array.Empty<Unit>();

	public MeasurementService(HttpClient _httpClient)
	{
		httpClient = _httpClient;
	}

	// <summary>
	// Fetches measurement data from the API.
	// </summary>
	public async Task<MeasurementCollection?> FetchMeasurementsAsync(int CollectionId)
	{
		// collection_id: int,
		// sample_id: int,
		// page_size: int,
		// page: int,
		// before: int, // UNIX timestamp in seconds
		// after: int // UNIX timestamp in seconds

		try
		{
			var response = await httpClient.GetAsync($"samples?collection_id={CollectionId}");
			if (response == null)
			{
				Console.WriteLine("Failed to fetch measurements: response is null");
				return null;
			}

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch measurements: {response.StatusCode}");
				return null;
			}

			string jsonString = await response.Content.ReadAsStringAsync();
			if (string.IsNullOrEmpty(jsonString))
			{
				Console.WriteLine("Failed to fetch measurements: response content is empty");
				return null;
			}

			var measurements = JsonSerializer.Deserialize<MeasurementCollection>(jsonString);
			if (measurements == null)
			{
				Console.WriteLine("Failed to deserialize measurements: response content is not valid JSON");
				return null;
			}

			if (units.Length == 0)
			{
				units = await FetchUnitsAsync();
			}
			// Insert the unit name into the headers
			foreach (var header in measurements.headers)
			{
				var unit = units.FirstOrDefault(u => u.Id == header.unitId);
				if (unit != null)
				{
					header.unit = unit.Name;
				}
			}

			return measurements;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching measurements: {ex.Message}");
			return null;
		}
	}

	// <summary>
	// Deletes a measurement by its ID.
	// </summary>
	public async Task<bool> DeleteMeasurementAsync(int sampleId)
	{
		try
		{
			HttpResponseMessage response = await httpClient.DeleteAsync($"samples?sample_id={sampleId}");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to delete measurement: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error deleting measurement: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Updates a measurement/sample.
	// </summary>
	public async Task<bool> UpdateMeasurementAsync(int sampleId, long? created_at, string? note)
	{
		// Query params:
		//	 sample_id: int,
		//	 created_at: int, // UNIX timestamp in seconds
		//	 note: string,
		if (created_at == null && note == null)
		{
			return false;
		}

		string qparams = "";
		if (created_at != null)
		{
			qparams += $"&created_at={created_at}";
		}
		if (note != null)
		{
			qparams += $"&note={Uri.EscapeDataString(note)}";
		}

		try
		{
			HttpResponseMessage response = await httpClient.PutAsync($"samples?sample_id={sampleId}{qparams}", new StringContent(""));

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to update measurement: {response.StatusCode} {response.Content.ReadAsStringAsync().Result}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error updating measurement: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Updates a value in a measurement.
	// </summary>
	public async Task<bool> UpdateMeasurementValueAsync(int sampleId, int attributeId, string value)
	{
		try
		{
			HttpResponseMessage response = await httpClient.PostAsync($"sample-values?sample_id={sampleId}&attribute_id={attributeId}&value={Uri.EscapeDataString(value)}", new StringContent(""));

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to update measurement: {response.StatusCode} {response.Content.ReadAsStringAsync().Result}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error updating measurement: {ex.Message}");
			return false;
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

	// <summary>
	// Adds a new measurement to a collection.
	// </summary>
	public async Task<bool> AddMeasurementAsync(int collectionId)
	{
		// 	{
		// 	collection_id: int,
		// 	note: string,
		// 	values: [
		// 		{
		// 			attribute_id: int,
		// 			value: string
		// 		}
		// 		...
		// 	]
		// }
		var content = new
		{
			collection_id = collectionId,
			values = Array.Empty<object>(),
		};
		try
		{
			HttpResponseMessage response = await httpClient.PostAsJsonAsync("samples", content);

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to add measurement: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error adding measurement: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Fetches all units.
	// </summary>
	public async Task<Unit[]> FetchUnitsAsync()
	{
		try
		{
			HttpResponseMessage response = await httpClient.GetAsync("units");

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to fetch units: {response.StatusCode}");
				return Array.Empty<Unit>();
			}

			string jsonString = await response.Content.ReadAsStringAsync();

			units = JsonSerializer.Deserialize<Unit[]>(jsonString) ?? Array.Empty<Unit>();
			return units;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error fetching units: {ex.Message}");
			return Array.Empty<Unit>();
		}
	}

	// <summary>
	// Add new attribute (column) to a collection.
	// </summary>
	public async Task<bool> AddAttributeAsync(int collectionId, string name, int? unitId)
	{

		// Add variables to query params
		var args = new Dictionary<string, string>
		{
			{ "collection_id", collectionId.ToString() },
			{ "name", name },
		};
		if (unitId != null)
		{
			args.Add("unit_id", unitId.ToString()!);
		}

		try
		{
			HttpResponseMessage response = await httpClient.PostAsync("attributes", new StringContent(new FormUrlEncodedContent(args).ReadAsStringAsync().Result, Encoding.UTF8, "application/x-www-form-urlencoded"));

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to add attribute: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error adding attribute: {ex.Message}");
			return false;
		}
	}

	// <summary>
	// Add a new unit.
	// </summary>
	public async Task<bool> AddUnitAsync(string name)
	{
		try
		{
			HttpResponseMessage response = await httpClient.PostAsync($"units?name={Uri.EscapeDataString(name)}", new StringContent(""));

			if (!response.IsSuccessStatusCode)
			{
				Console.WriteLine($"Failed to add unit: {response.StatusCode}");
				return false;
			}
			return true;
		}
		catch (Exception ex)
		{
			Console.WriteLine($"Error adding unit: {ex.Message}");
			return false;
		}
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

public class Unit
{
	[JsonPropertyName("id")]
	public int Id { get; set; }

	[JsonPropertyName("name")]
	public string Name { get; set; } = "";
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
	[JsonPropertyName("attributes")]
	public List<MeasurementHeader> headers { get; set; } = new();

	[JsonPropertyName("samples")]
	public List<Measurement> entries { get; set; } = new();

	[JsonPropertyName("total_count")]
	public int TotalCount { get; set; }

	[JsonPropertyName("page_info")]
	public PageInfo PageInfo { get; set; } = new();
}

public class MeasurementHeader
{
	[JsonPropertyName("attribute_id")]
	public int Id { get; set; }

	public string name { get; set; } = string.Empty;

	[JsonPropertyName("unit_id")]
	public int unitId { get; set; }

	public string unit { get; set; } = string.Empty;
}
public class Measurement
{
	[JsonPropertyName("sample_id")]
	public int Id { get; set; }

	[JsonPropertyName("note")]
	public string Note { get; set; } = string.Empty;

	[JsonPropertyName("created_at")]
	public long CreatedAt { get; set; }

	public List<MeasurementValue> values { get; set; } = new();
}
public class MeasurementValue
{
	[JsonPropertyName("attribute_id")]
	public int Id { get; set; }

	public string value { get; set; } = string.Empty;
}
public class PageInfo
{
	public int Limit { get; set; }
	public int Offset { get; set; }

	[JsonPropertyName("has_next_page")]
	public bool HasNextPage { get; set; }
}