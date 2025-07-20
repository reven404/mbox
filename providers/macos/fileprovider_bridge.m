#import <Foundation/Foundation.h>
#import <FileProvider/FileProvider.h>
#import "fileprovider_bridge.h"

// Global domain manager
static NSMutableDictionary<NSString *, NSFileProviderDomain *> *activeDomains = nil;
static dispatch_once_t onceToken;

// Initialize domain storage
static void initialize_domains(void) {
    dispatch_once(&onceToken, ^{
        activeDomains = [[NSMutableDictionary alloc] init];
    });
}

// Convert C string to NSString
static NSString *cstring_to_nsstring(const char *cstr) {
    if (cstr == NULL) return nil;
    return [NSString stringWithUTF8String:cstr];
}

// Convert NSString to C string (caller must free)
static char *nsstring_to_cstring(NSString *nsstr) {
    if (nsstr == nil) return NULL;
    const char *utf8 = [nsstr UTF8String];
    size_t len = strlen(utf8) + 1;
    char *cstr = malloc(len);
    if (cstr) {
        strcpy(cstr, utf8);
    }
    return cstr;
}

// Register a new File Provider domain
int fp_register_domain(const char* domain_id, const char* display_name, const char* server_url) {
    @autoreleasepool {
        initialize_domains();
        
        NSString *domainID = cstring_to_nsstring(domain_id);
        NSString *displayName = cstring_to_nsstring(display_name);
        NSString *serverURL = cstring_to_nsstring(server_url);
        
        if (!domainID || !displayName) {
            return -1; // Invalid parameters
        }
        
        // Create File Provider domain
        NSFileProviderDomain *domain = [[NSFileProviderDomain alloc] 
            initWithIdentifier:domainID 
            displayName:displayName];
        
        // Store server URL in user defaults for the extension to read
        NSUserDefaults *defaults = [[NSUserDefaults alloc] 
            initWithSuiteName:@"group.com.mdriver.fileprovider"];
        if (serverURL) {
            [defaults setObject:serverURL forKey:[NSString stringWithFormat:@"%@.serverURL", domainID]];
        }
        [defaults synchronize];
        
        // Add domain to File Provider manager (macOS uses different API)
        [NSFileProviderManager addDomain:domain completionHandler:^(NSError * _Nullable error) {
            if (error) {
                NSLog(@"Failed to add File Provider domain: %@", error.localizedDescription);
            } else {
                NSLog(@"Successfully registered File Provider domain: %@", domainID);
            }
        }];
        
        // Store in our local registry
        activeDomains[domainID] = domain;
        
        return 0; // Success
    }
}

// Remove a File Provider domain
int fp_remove_domain(const char* domain_id) {
    @autoreleasepool {
        initialize_domains();
        
        NSString *domainID = cstring_to_nsstring(domain_id);
        if (!domainID) {
            return -1;
        }
        
        NSFileProviderDomain *domain = activeDomains[domainID];
        if (!domain) {
            return -2; // Domain not found
        }
        
        // Remove from File Provider manager
        [NSFileProviderManager removeDomain:domain completionHandler:^(NSError * _Nullable error) {
            if (error) {
                NSLog(@"Failed to remove File Provider domain: %@", error.localizedDescription);
            } else {
                NSLog(@"Successfully removed File Provider domain: %@", domainID);
            }
        }];
        
        // Clean up user defaults
        NSUserDefaults *defaults = [[NSUserDefaults alloc] 
            initWithSuiteName:@"group.com.mdriver.fileprovider"];
        [defaults removeObjectForKey:[NSString stringWithFormat:@"%@.serverURL", domainID]];
        [defaults synchronize];
        
        // Remove from local registry
        [activeDomains removeObjectForKey:domainID];
        
        return 0;
    }
}

// Signal that enumeration has changed for a container
int fp_signal_enumeration_change(const char* domain_id, const char* container_id) {
    @autoreleasepool {
        NSString *domainID = cstring_to_nsstring(domain_id);
        NSString *containerID = cstring_to_nsstring(container_id);
        
        if (!domainID) return -1;
        
        NSFileProviderDomain *domain = activeDomains[domainID];
        if (!domain) return -2;
        
        NSFileProviderManager *manager = [NSFileProviderManager managerForDomain:domain];
        NSFileProviderItemIdentifier identifier = containerID ? 
            containerID : 
            NSFileProviderRootContainerItemIdentifier;
        
        [manager signalEnumeratorForContainerItemIdentifier:identifier completionHandler:^(NSError * _Nullable error) {
            if (error) {
                NSLog(@"Failed to signal enumeration change: %@", error.localizedDescription);
            }
        }];
        
        return 0;
    }
}

// Signal that a specific item has changed
int fp_signal_item_change(const char* domain_id, const char* item_id) {
    @autoreleasepool {
        NSString *domainID = cstring_to_nsstring(domain_id);
        NSString *itemID = cstring_to_nsstring(item_id);
        
        if (!domainID || !itemID) return -1;
        
        NSFileProviderDomain *domain = activeDomains[domainID];
        if (!domain) return -2;
        
        NSFileProviderManager *manager = [NSFileProviderManager managerForDomain:domain];
        NSFileProviderItemIdentifier identifier = itemID;
        
        // Signal that the item changed
        [manager signalEnumeratorForContainerItemIdentifier:identifier completionHandler:^(NSError * _Nullable error) {
            if (error) {
                NSLog(@"Failed to signal item change: %@", error.localizedDescription);
            }
        }];
        
        return 0;
    }
}

// Set server URL for a domain
int fp_set_server_url(const char* domain_id, const char* server_url) {
    @autoreleasepool {
        NSString *domainID = cstring_to_nsstring(domain_id);
        NSString *serverURL = cstring_to_nsstring(server_url);
        
        if (!domainID) return -1;
        
        NSUserDefaults *defaults = [[NSUserDefaults alloc] 
            initWithSuiteName:@"group.com.mdriver.fileprovider"];
        
        if (serverURL) {
            [defaults setObject:serverURL forKey:[NSString stringWithFormat:@"%@.serverURL", domainID]];
        } else {
            [defaults removeObjectForKey:[NSString stringWithFormat:@"%@.serverURL", domainID]];
        }
        [defaults synchronize];
        
        return 0;
    }
}

// Get server URL for a domain
char* fp_get_server_url(const char* domain_id) {
    @autoreleasepool {
        NSString *domainID = cstring_to_nsstring(domain_id);
        if (!domainID) return NULL;
        
        NSUserDefaults *defaults = [[NSUserDefaults alloc] 
            initWithSuiteName:@"group.com.mdriver.fileprovider"];
        NSString *serverURL = [defaults stringForKey:[NSString stringWithFormat:@"%@.serverURL", domainID]];
        
        return nsstring_to_cstring(serverURL);
    }
}

// Enable or disable a domain
int fp_set_domain_enabled(const char* domain_id, int enabled) {
    @autoreleasepool {
        NSString *domainID = cstring_to_nsstring(domain_id);
        if (!domainID) return -1;
        
        NSUserDefaults *defaults = [[NSUserDefaults alloc] 
            initWithSuiteName:@"group.com.mdriver.fileprovider"];
        [defaults setBool:(enabled != 0) forKey:[NSString stringWithFormat:@"%@.enabled", domainID]];
        [defaults synchronize];
        
        return 0;
    }
}

// Get domain status
int fp_get_domain_status(const char* domain_id) {
    @autoreleasepool {
        initialize_domains();
        
        NSString *domainID = cstring_to_nsstring(domain_id);
        if (!domainID) return -1;
        
        NSFileProviderDomain *domain = activeDomains[domainID];
        if (!domain) return 0; // Not registered
        
        return 1; // Registered
    }
}

// Get domain error (if any)
char* fp_get_domain_error(const char* domain_id) {
    @autoreleasepool {
        // For now, return NULL (no error tracking implemented)
        return NULL;
    }
}

// Free C string allocated by this module
void fp_free_string(char* str) {
    if (str) {
        free(str);
    }
}