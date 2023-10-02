//
// Copyright © Mark Burgess, ChiTek-i (2023)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// To run, install package:
//
//      go get github.com/sashabaranov/go-openai
//      go run chatgpt_client.go
//
// This is tuned specifically to Wikipedia scanning. using the general methods
// ***********************************************************

package main

import (
	"context"
	"fmt"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

// *************************************************************************

const QUERY = "Summarize feelings of trust in this text: \"The more or less automatically added ISSUES macros have been deleted. A full, thorough update of the article fixing all the mentioned issues was rejected by an editor, so the complaints must have been a mistake. Hazitt (talk) 16:52, 12 January 2023 (UTC) The templates were not automatically added, they were added after an editor with a clear conflict of interest edited the article and introduced the issues. The purpose of maintenance templates is to encourage other editors to come and work on the article to address them, but unfprauntegly in this case they have not attracted much attention in the last 6 years. Your edit did not address all the issues because it added large sections of unsourced content, and content that reads like original research rather than being based on independent, secondary sources. Do you have a connection to Mark Burgess yourself? Thank you Melcous (talk) 21:46, 12 January 2023 (UTC) Melcous, that is a bold accusation indeed: 'with a clear conflict of interest'. If this is how you work - throwing accusations around in order to qualify questionable decisions - I sincerely doubt you are the right person for the job. Do you actually understand the subject at hand at all? I suggest a more cooperative attitude starting with a well researched and referenced article and add/edit/fix whatever is not within the rules. Randomly deleting paragraphs and sentences creates a nonsensical result, making the article worse than worthless. You should know that. This is a complex, theoretical subject, take a look at Quantum field theory and see how far you get with a paragraph or 5. It just doesn't work that way with scientific disciplines. Hazitt (talk) 16:40, 13 January 2023 (UTC) Hazitt it is not bold, nor is it an accusation; it is a simple statement of fact. The editor I referred to (after whose editing the templates were added) acknowledged here exactly who they are, and therefore that they have a \"clear conflict of interest\". Perhaps you should read WP:AGF. Melcous (talk) 21:34, 17 January 2023 (UTC) Seriously, are you really referring to an exchange from 2016 in order to reject a rewrite by someone else in 2023? What in the world do I have to do with that? This is insane. There is just too much that doesn't make sense in your responses. You are creating 'facts' from thin air and seem to be believing in them. I think it's time you explain yourself. What are you trying to achieve here, other than torpedo some serious work to enhance wikipedia? For the record - where is the unsourced content in my contribution? Did you care to check the references? It's all there if you'd cared to look. You're obviously not qualified to deal with this so why don't you just recuse yourself and let's move on? Hazitt (talk) 21:48, 20 January 2023 (UTC) Hazitt please try to comment on content not contributors. I have tried to carefully respond to what you have said - for example, you started by saying the templates were added \"automatically\" and \"in error\". I explained that that was not the case. You then said I was making accusations. I explained why that was an incorrect characterisation. Now you are asking about unsourced content? The current maintenance templates on the article have nothing to do with content being unsourced. They flag the need for independent, secondary sources and suggest there is content that reads as original research (i.e. an editor's interpretation of primary sources). Melcous (talk) 11:12, 22 January 2023 (UTC) Melcous, you call me \"an editor with a clear conflict of interest \" - that's a bold accusation. If you're referring to older edits by a different editor, you need to point that out, get the timeline right. You removed a paragraph on Jan 12th@21:42 and called it 'unsourced'. You slam on WP macros seemingly without even trying to get the whole picture (I call that 'more or less automatically' because there does not seem to be much reflection being the actions). Please read the WP definition of original research again, then identify where you find this. We can take it from there. Hazitt (talk) 13:36, 31 January 2023 (UTC) Hazitt This is quickly turning into a fruitless back and forth. But to be clear again, I did not call you \"an editor with a clear conflict of interest\" - I explained that the maintenance templates were added after a COI editor edited the article, and it is a simple matter of looking at the edit history to see that this was in August 2016 after MarkBurgess edited the article. The maintenance templates also clearly had the date August 2016 on them, plus my comment here explicitly noted that it had been 6 years since that occurred - which was well before you had ever edited here. That you chose to misinterpret that as referring to you is not something I could have predicted. Yes, I removed a paragraph (well, technically 3 paragraphs) on 12 Jan here and called it 'unsourced', because it was entirely unsourced, i.e. it had no references. And thank you for your suggestions, but I am very familiar with the definition of WP:OR and it includes \"any analysis or synthesis of published material that serves to reach or imply a conclusion not stated by the sources\". Once again, I suggest you try commenting on content not contributors. Melcous (talk) 13:56, 31 January 2023 (UTC) Well this is interesting. I just got flagged by email about this discussion. So a brief comment. Yes, I did the original work on Promise Theory. i) I'm happy to see someone trying to clean up this article. It was pretty awful and misleading, but I have to say looking better than before. ii) I have some sympathy for Wikipedia's position on people promoting themselves, but I got slapped once before for trying to correct mistakes people had written, and so I am keeping out of it. However, there are clear double standards here. Melcous, do you have some grudge against me personally? I don't know who you are. You seem pretty determined to stop any improvement of this page. What's your grudge? Perhaps you should pass this to a different editor to deal with. I have given up trying to contribute to Wikipedia because of this kind of thing. Over and out. MarkBurgess (talk) 14:33, 31 January 2023 (UTC) MarkBurgess, I will not tag you again after this as it seems you do not want to be involved (or you can turn off your email notifications). I referenced your name in an admittedly long-winded explanation about why maintenance tags were added to this article that another editor seems to have misinterpreted. I have no grudge against you, I don't really know who you are either. I am not trying to stop improvement of this or any other page, but I am trying to abide by wikipedia's core guidelines which include no original research or editing to promote something - hence the conflict of interest guidelines which you were pointed to previously. I do not think this was at all \"slapping\" you and in fact, you and I had a reasonably polite exchange at the time where I explained the no original research policy and you thanked me for a \"sympathetic ear\". Unfortunately, since then, there have not been editors who have been able to provide reliable, independent secondary sources on this topic, which is what has led to this being discussed again. Cheers, Melcous (talk) 14:45, 31 January 2023 (UTC) Melcous, I agree - this is fruitless. We cannot even agree on what happened when even though the evidence is open and available in the article history. For the record let me point out that three or more paragraphs without a reference is the normal, not the exception for scientific articles. Again I suggest that you read and at least try to understand the content before deleting - or marking something as unsourced or ':OR'. Hazitt (talk) 11:34, 13 February 2023 (UTC) Hazitt can you please point me to a wikipedia policy or guidelines that says that three or more paragraphs without a reference is acceptable in scientific articles, and how it is different to WP:BURDEN? Thank you Melcous (talk) 11:43, 13 February 2023 (UTC) Thank you @Melcous, that explains a lot. You're stickler for rules. Fortunately, Wikipedia is not - if it were, it would not exist. Wikipedia's success is based on a healthy combination of rules and common sense. You'll find scores of scientific (and other types of) articles with many consecutive paragraphs without (new) citations, acknowledging the fact that repeating the same reference again and again is noise and nothing else. BTW, WP:BURDEN is pragmatically openminded about this, as it should be. Hazitt (talk) 12:58, 23 February 2023 (UTC)\""

// *************************************************************************

func main() {

	client := openai.NewClient(os.Getenv("API_KEY"))

	resp, err := client.CreateChatCompletion(

		context.Background(),

		openai.ChatCompletionRequest{

			Model: openai.GPT3Dot5Turbo,

			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: QUERY,

				},
			},
		},
	)

	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return
	}

	fmt.Println(resp.Choices[0].Message.Content)
}