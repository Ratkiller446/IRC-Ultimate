package art

import (
	"fmt"
	"strings"
)

// CatFaces contains various kawaii cat faces using ASCII characters
var CatFaces = []string{
	`  /\_/\  
 ( o.o ) 
  > ^ <  `,
	
	`  /\_/\  
 ( -.- ) 
  ( >< ) `,
	
	`  /\_/\  
 ( ^.^ ) 
  (") (") `,
	
	`  /\_/\  
 ( ='.'=)
  (")_(") `,
	
	`  /\_/\  
 ( ^.^ ) 
  (("))  `,
}

// GetCatWithMessage creates a cat with a message bubble
func GetCatWithMessage(msg string) string {
	return fmt.Sprintf(
		`  /\_/\    
 ( o.o )   %s
  > ^ <    %s`,
		"<"+strings.Repeat("-", len(msg)+2)+">",
		"( "+msg+" )",
	)
}
