#!/usr/bin/env python3
from owslib.ogcapi.features import Features

API_ENDPOINT = 'http://localhost:8080'

print(f"Attempting to connect to OGC API at: {API_ENDPOINT}")

try:
    ogc_api = Features(API_ENDPOINT)

    collections = ogc_api.feature_collections()
    print(f"\nSuccessfully retrieved collections:")
    for collection in collections:
        print(collection)

    collection_id = 'addresses'
    print(f"\nFetching up to 5 features from the '{collection_id}' collection...")
    retrieved_features = ogc_api.collection_items(collection_id, limit='5')

    print(f"Found {len(retrieved_features['features'])} features:")
    for feature in retrieved_features['features']:
        feature_name = feature.get('properties', {}).get('component_postaldescriptor', 'N/A')
        print(f"  - Feature ID: {feature['id']}, Postal code: {feature_name}")

    print("\n Done.")

except Exception as e:
    print(f"\n An error occurred: {e}")
    exit(1)